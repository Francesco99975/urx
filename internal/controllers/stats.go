package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Francesco99975/urx/cmd/boot"
	"github.com/Francesco99975/urx/internal/config"
	"github.com/Francesco99975/urx/internal/database"
	"github.com/Francesco99975/urx/internal/helpers"
	"github.com/Francesco99975/urx/internal/repository"
	"github.com/Francesco99975/urx/views"
	"github.com/dustin/go-humanize"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Stats() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := config.GetDefaultSite(c.Request())

		ctx := c.Request().Context()

		tx, err := database.Pool().BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error")
		}

		defer database.HandleTransaction(ctx, tx, &err)

		repo := repository.New(tx)

		publicGlobalLinksAmount, err := repo.CountPublicLinks(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not count public global links")
		}

		publicGlobalClicksAmount, err := repo.SumPublicClicks(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not sum public global clicks")
		}

		publicGlobalClicksAmountToday, err := repo.SumPublicClicksToday(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not sum public global clicks today")
		}

		publicGlobalClicksAmountWeek, err := repo.SumPublicClicksLast7Days(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not sum public global clicks week")
		}

		weekChange := float64(publicGlobalClicksAmountWeek) / float64(publicGlobalClicksAmount) * 100

		sign := "+"

		if weekChange < 0 {
			sign = "-"
		}

		mainStats := views.DashboardStats{
			TotalLinks:   strconv.FormatInt(publicGlobalLinksAmount, 10),
			TotalClicks:  strconv.FormatInt(publicGlobalClicksAmount, 10),
			WeeklyChange: fmt.Sprintf("%s%s%s", sign, strconv.FormatFloat(weekChange, 'f', 2, 64), "%"),
			ClicksToday:  strconv.FormatInt(publicGlobalClicksAmountToday, 10),
		}

		log.Debugf("Total links: %s", mainStats.TotalLinks)
		log.Debugf("Total clicks: %s", mainStats.TotalClicks)
		log.Debugf("Weekly change: %s", mainStats.WeeklyChange)
		log.Debugf("Clicks today: %s", mainStats.ClicksToday)

		publicGlobalVisitedUrls, err := repo.GetTopPublicURLs(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not get public global visited urls")
		}

		var topLinks []views.TopLink

		for i, link := range publicGlobalVisitedUrls {
			if link.TotalClicks == 0 {
				continue
			}

			trend := link.WeekClicks / link.TotalClicks * 100
			var trendClass string
			var sign string
			if trend > 0 {
				trendClass = "text-success"
				sign = "+"
			} else if trend < 0 {
				trendClass = "text-error"
				sign = "-"
			} else {
				trendClass = "text-gray-500"
			}

			topLinks = append(topLinks, views.TopLink{
				Rank:        i + 1,
				ShortURL:    fmt.Sprintf("%s/%s", boot.Environment.URL, link.Code),
				OriginalURL: link.LongUrl,
				StatsURL:    "",
				Clicks:      strconv.FormatInt(link.TotalClicks, 10),
				Trend:       fmt.Sprintf("%s%s%s", sign, strconv.FormatInt(trend, 10), "%"),
				TrendClass:  trendClass,
			})
		}

		log.Debugf("Top links: %v", topLinks)

		publicGlobalReferrers, err := repo.GetTopPublicReferrers(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not get public global referrers")
		}

		var topReferrers []views.ReferrerStat

		for _, referrer := range publicGlobalReferrers {
			percent := float64(referrer.Clicks) / float64(publicGlobalClicksAmount) * 100
			letter, bg, text := getKnowReferers(referrer.Referer)

			topReferrers = append(topReferrers, views.ReferrerStat{
				Name:     referrer.Referer,
				Count:    strconv.FormatInt(referrer.Clicks, 10),
				Percent:  int(percent),
				Letter:   letter,
				BgClass:  bg,
				TxtClass: text,
				IsGlobe:  referrer.Referer == "Direct",
			})
		}

		log.Debugf("Top referrers: %v", topReferrers)

		deviceBreakdown, err := repo.GetPublicDeviceBreakdown(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not get device breakdown")
		}

		var deviceStats []views.DeviceStat

		for _, device := range deviceBreakdown {
			var name string
			if device.Device == nil {
				name = "unknown"
			} else {
				if *device.Device == "" {
					continue
				}

				name = *device.Device
			}

			percentage := float64(device.Clicks) / float64(publicGlobalClicksAmount) * 100

			deviceStats = append(deviceStats, views.DeviceStat{
				Name:       helpers.Capitalize(name),
				DeviceType: name,
				PercentInt: int(percentage),
				Percent:    fmt.Sprintf("%s%s", strconv.FormatInt(int64(percentage), 10), "%"),
			})
		}

		log.Debugf("Device stats: %v", deviceStats)

		browserBreakdown, err := repo.GetPublicBrowserBreakdown(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not get browser breakdown")
		}

		var browserStats []views.BrowserStat

		for _, browser := range browserBreakdown {
			percentage := float64(browser.Clicks) / float64(publicGlobalClicksAmount) * 100
			var label string
			if browser.Browser == nil {
				label = fmt.Sprintf("%s %s%s", "Unknown", strconv.FormatInt(int64(percentage), 10), "%")
			} else {
				label = fmt.Sprintf("%s %s%s", helpers.Capitalize(*browser.Browser), strconv.FormatInt(int64(percentage), 10), "%")
			}

			browserStats = append(browserStats, views.BrowserStat{
				Label: label,
			})
		}

		log.Debugf("Browser stats: %v", browserStats)

		publicRecentClicks, err := repo.GetRecentPublicClicks(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Server Error could not get recent clicks")
		}

		var activities []views.ActivityItem

		for _, activity := range publicRecentClicks {
			var device string
			if activity.Browser != nil && activity.Os != nil {
				device = fmt.Sprintf("%s / %s", helpers.Capitalize(*activity.Browser), helpers.Capitalize(*activity.Os))
			} else {
				device = "Unknown"
			}

			activities = append(activities, views.ActivityItem{
				Time:     humanize.Time(activity.CreatedAt.Time),
				ShortURL: fmt.Sprintf("%s/%s", boot.Environment.URL, activity.Code),
				StatsURL: "",
				Device:   device,
				Referrer: activity.Referer,
				IsDirect: activity.Referer == "Direct",
			})
		}

		log.Debugf("Activities: %v", activities)

		statsParams := views.StatsPageData{
			Stats:      mainStats,
			TopLinks:   topLinks,
			Referrers:  topReferrers,
			Devices:    deviceStats,
			Browsers:   browserStats,
			Activities: activities,
		}

		data.CSRF = c.Get("csrf").(string)
		data.Nonce = c.Get("nonce").(string)

		html := helpers.MustRenderHTML(views.StatsPage(data, statsParams))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}

func getKnowReferers(referer string) (string, string, string) {
	if strings.Contains(referer, "google") {
		return "G", "bg-red", "text-red"
	}

	if strings.Contains(referer, "facebook") {
		return "F", "bg-blue", "text-blue"
	}

	if strings.Contains(referer, "twitter") {
		return "T", "bg-cyan", "text-cyan"
	}

	if strings.Contains(referer, "x") {
		return "T", "bg-cyan", "text-cyan"
	}

	if strings.Contains(referer, "youtube") {
		return "Y", "bg-red", "text-red"
	}

	if strings.Contains(referer, "twitch") {
		return "W", "bg-pink", "text-pink"
	}

	if strings.Contains(referer, "instagram") {
		return "I", "bg-violet", "text-violet"
	}

	if strings.Contains(referer, "linkedin") {
		return "L", "bg-indigo", "text-indigo"
	}

	if strings.Contains(referer, "pinterest") {
		return "P", "bg-pink", "text-pink"
	}

	if strings.Contains(referer, "reddit") {
		return "R", "bg-orange", "text-orange"
	}

	return "O", "bg-gray", "text-gray"
}
