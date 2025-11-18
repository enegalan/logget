package core

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	chrome "logget/src/chrome"
	"github.com/go-rod/rod/lib/proto"
)

func SetCookies(ctx *chrome.ChromeContext, targetURL string, cookies []string) error {
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
	rodCookies := []*proto.NetworkCookie{}
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
	var sameSite proto.NetworkCookieSameSite
	var expires *proto.TimeSinceEpoch
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
				sameSiteStr := strings.ToLower(strings.TrimSpace(attr[9:]))
				switch sameSiteStr {
				case "strict":
					sameSite = proto.NetworkCookieSameSiteStrict
				case "lax":
					sameSite = proto.NetworkCookieSameSiteLax
				case "none":
					sameSite = proto.NetworkCookieSameSiteNone
				}
			case strings.HasPrefix(attrLower, "expires="):
				expiresStr := strings.TrimSpace(attr[8:])
				if parsedExpires, err := time.Parse(time.RFC1123, expiresStr); err == nil {
					expiresTS := proto.TimeSinceEpoch(parsedExpires.Unix())
					expires = &expiresTS
				} else if parsedExpires, err := time.Parse(time.RFC1123Z, expiresStr); err == nil {
					expiresTS := proto.TimeSinceEpoch(parsedExpires.Unix())
					expires = &expiresTS
				} else {
					return fmt.Errorf("invalid expires format: %s", expiresStr)
				}
			}
		}
		cookie := &proto.NetworkCookie{
			Name:     name,
			Value:    value,
			Domain:   cookieDomain,
			Path:     path,
			Secure:   secure,
			HTTPOnly: httpOnly,
		}
		if sameSite != "" {
			cookie.SameSite = sameSite
		}
		if expires != nil {
			cookie.Expires = proto.TimeSinceEpoch(*expires)
		}
		rodCookies = append(rodCookies, cookie)
	}
	if len(rodCookies) > 0 {
		for _, cookie := range rodCookies {
			ctx.Page.MustSetCookies(&proto.NetworkCookieParam{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Domain:   cookie.Domain,
				Path:     cookie.Path,
				Secure:   cookie.Secure,
				HTTPOnly: cookie.HTTPOnly,
				SameSite: cookie.SameSite,
				Expires:  cookie.Expires,
			})
		}
	}
	return nil
}
