package helpers

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	cdpnetwork "github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func SetCookies(ctx context.Context, targetURL string, cookies []string) error {
	if len(cookies) == 0 {
		return nil
	}
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}
	domain := parsedURL.Host
	if !strings.Contains(domain, ".") {
		domain = "." + domain
	}
	for _, cookieStr := range cookies {
		parts := strings.Split(cookieStr, ";")
		nameValue := strings.TrimSpace(parts[0])
		if !strings.Contains(nameValue, "=") {
			return fmt.Errorf("invalid cookie format: %s (expected 'name=value')", cookieStr)
		}
		nameValueParts := strings.SplitN(nameValue, "=", 2)
		name := strings.TrimSpace(nameValueParts[0])
		value := strings.TrimSpace(nameValueParts[1])
		cookieDomain := domain
		path := "/"
		secure := parsedURL.Scheme == "https"
		httpOnly := false
		sameSite := ""
		var expires *time.Time
		for i := 1; i < len(parts); i++ {
			attr := strings.TrimSpace(parts[i])
			attrLower := strings.ToLower(attr)
			switch {
			case strings.HasPrefix(attrLower, "domain="):
				cookieDomain = strings.TrimSpace(attr[7:])
			case strings.HasPrefix(attrLower, "path="):
				path = strings.TrimSpace(attr[5:])
			case attrLower == "secure":
				secure = true
			case attrLower == "httponly":
				httpOnly = true
			case strings.HasPrefix(attrLower, "samesite="):
				sameSite = strings.TrimSpace(attr[9:])
			case strings.HasPrefix(attrLower, "expires="):
				expiresStr := strings.TrimSpace(attr[8:])
				if parsedExpires, err := time.Parse(time.RFC1123, expiresStr); err == nil {
					expires = &parsedExpires
				} else if parsedExpires, err := time.Parse(time.RFC1123Z, expiresStr); err == nil {
					expires = &parsedExpires
				} else {
					return fmt.Errorf("invalid expires format: %s", expiresStr)
				}
			}
		}
		cookieCmd := cdpnetwork.SetCookie(name, value).WithDomain(cookieDomain).WithPath(path).WithSecure(secure).WithHTTPOnly(httpOnly)
		if sameSite != "" {
			switch strings.ToLower(sameSite) {
			case "strict":
				cookieCmd = cookieCmd.WithSameSite(cdpnetwork.CookieSameSiteStrict)
			case "lax":
				cookieCmd = cookieCmd.WithSameSite(cdpnetwork.CookieSameSiteLax)
			case "none":
				cookieCmd = cookieCmd.WithSameSite(cdpnetwork.CookieSameSiteNone)
			}
		}
		if expires != nil {
			expiresTime := cdp.TimeSinceEpoch(*expires)
			cookieCmd = cookieCmd.WithExpires(&expiresTime)
		}
		if err := chromedp.Run(ctx, cookieCmd); err != nil {
			return fmt.Errorf("failed to set cookie %s: %v", name, err)
		}
	}
	return nil
}
