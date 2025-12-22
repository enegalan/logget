package core

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	chrome "logget/src/chrome"
)

type fingerprintConfig struct {
	UserAgent           string   `json:"userAgent"`
	Platform            string   `json:"platform"`
	Language            string   `json:"language"`
	Languages           []string `json:"languages"`
	ScreenWidth         int      `json:"screenWidth"`
	ScreenHeight        int      `json:"screenHeight"`
	ColorDepth          int      `json:"colorDepth"`
	PixelDepth          int      `json:"pixelDepth"`
	HardwareConcurrency int      `json:"hardwareConcurrency"`
	DeviceMemory        int      `json:"deviceMemory"`
	MaxTouchPoints      int      `json:"maxTouchPoints"`
	WebGLVendor         string   `json:"webglVendor"`
	WebGLRenderer       string   `json:"webglRenderer"`
}

func RotateNavigatorFingerprints(ctx *chrome.ChromeContext, intervalMs int) error {
	fingerprint := generateFingerprint()
	fingerprintJSON, err := json.Marshal(fingerprint)
	if err != nil {
		return fmt.Errorf("failed to marshal fingerprint: %v", err)
	}
	js := fmt.Sprintf(`() => {
		try {
			if (typeof Object === 'undefined' || typeof Object.defineProperty !== 'function') return;
			const fingerprint = %s;
			if (!fingerprint || typeof fingerprint !== 'object') return;
			if (typeof navigator === 'undefined' || navigator === null) return;
			if (typeof window !== 'undefined') {
				if (!window.__loggetFingerprintValues) {
					window.__loggetFingerprintValues = {};
				}
				var values = window.__loggetFingerprintValues;
				
				if (fingerprint.userAgent) values.userAgent = fingerprint.userAgent;
				if (fingerprint.platform) values.platform = fingerprint.platform;
				if (fingerprint.language) values.language = fingerprint.language;
				if (fingerprint.languages) values.languages = fingerprint.languages;
				if (typeof fingerprint.hardwareConcurrency !== 'undefined') values.hardwareConcurrency = fingerprint.hardwareConcurrency;
				if (typeof fingerprint.deviceMemory !== 'undefined') values.deviceMemory = fingerprint.deviceMemory;
				if (typeof fingerprint.maxTouchPoints !== 'undefined') values.maxTouchPoints = fingerprint.maxTouchPoints;
				if (typeof fingerprint.screenWidth !== 'undefined') values.screenWidth = fingerprint.screenWidth;
				if (typeof fingerprint.screenHeight !== 'undefined') values.screenHeight = fingerprint.screenHeight;
				if (typeof fingerprint.colorDepth !== 'undefined') values.colorDepth = fingerprint.colorDepth;
				if (typeof fingerprint.pixelDepth !== 'undefined') values.pixelDepth = fingerprint.pixelDepth;
				function safeDefineProperty(obj, prop, valueKey) {
					try {
						if (!obj || typeof obj !== 'object' || obj === null) return;
						if (!prop || typeof prop !== 'string') return;
						if (!valueKey || typeof valueKey !== 'string') return;
						if (!values || typeof values !== 'object') return;
						if (typeof values[valueKey] === 'undefined') return;
						
						var val = values[valueKey];
						if (val === undefined || val === null) return;
						try {
							Object.defineProperty(obj, prop, {
								get: function() { return val; },
								configurable: true
							});
						} catch(e) {
							// Ignore - property may not be configurable
						}
					} catch(e) {
						// Ignore all errors
					}
				}
				try { // Override navigator.userAgent
					if (values.userAgent) safeDefineProperty(navigator, 'userAgent', 'userAgent');
				} catch(e) {}
				try { // Override navigator.platform
					if (values.platform) safeDefineProperty(navigator, 'platform', 'platform');
				} catch(e) {}
				try { // Override navigator.language
					if (values.language) safeDefineProperty(navigator, 'language', 'language');
				} catch(e) {}
				try { // Override navigator.languages
					if (values.languages) safeDefineProperty(navigator, 'languages', 'languages');
				} catch(e) {}
				try { // Override navigator.hardwareConcurrency
					if (typeof values.hardwareConcurrency !== 'undefined') safeDefineProperty(navigator, 'hardwareConcurrency', 'hardwareConcurrency');
				} catch(e) {}
				try { // Override navigator.deviceMemory
					if (typeof navigator !== 'undefined' && 'deviceMemory' in navigator && typeof values.deviceMemory !== 'undefined') {
						safeDefineProperty(navigator, 'deviceMemory', 'deviceMemory');
					}
				} catch(e) {}
				try { // Override navigator.maxTouchPoints
					if (typeof values.maxTouchPoints !== 'undefined') safeDefineProperty(navigator, 'maxTouchPoints', 'maxTouchPoints');
				} catch(e) {}
				try { // Override screen properties
					if (typeof screen !== 'undefined') {
						if (typeof values.screenWidth !== 'undefined') safeDefineProperty(screen, 'width', 'screenWidth');
						if (typeof values.screenHeight !== 'undefined') safeDefineProperty(screen, 'height', 'screenHeight');
						if (typeof values.screenWidth !== 'undefined') safeDefineProperty(screen, 'availWidth', 'screenWidth');
						if (typeof values.screenHeight !== 'undefined') {
							try {
								var availH = values.screenHeight - 40;
								Object.defineProperty(screen, 'availHeight', {
									get: function() { return availH; },
									configurable: true
								});
							} catch(e) {}
						}
						if (typeof values.colorDepth !== 'undefined') safeDefineProperty(screen, 'colorDepth', 'colorDepth');
						if (typeof values.pixelDepth !== 'undefined') safeDefineProperty(screen, 'pixelDepth', 'pixelDepth');
					}
				} catch(e) {}
			}
		} catch(e) {
			// Ignore all errors
		}
	}`, string(fingerprintJSON))
	_, err = ctx.Page.Eval(js)
	if err != nil {
		return fmt.Errorf("failed to inject fingerprint rotation script: %v", err)
	}
	return nil
}

func StartFingerprintRotation(ctx *chrome.ChromeContext, intervalMs int) error {
	if intervalMs <= 0 {
		return nil
	}
	if err := RotateNavigatorFingerprints(ctx, intervalMs); err != nil { // Initial rotation
		return err
	}
	go func() { // Set up periodic rotation
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Ctx.Done():
				return
			case <-ticker.C:
				if err := RotateNavigatorFingerprints(ctx, intervalMs); err != nil {
					continue
				}
			}
		}
	}()
	return nil
}

func generateFingerprint() fingerprintConfig {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	}
	platforms := []string{
		"Win32",
		"MacIntel",
		"Linux x86_64",
	}
	languages := []string{
		"en-US",
		"en-GB",
		"en-CA",
		"en-AU",
		"fr-FR",
		"fr-CA",
		"de-DE",
		"es-ES",
		"es-MX",
		"it-IT",
		"pt-BR",
		"pt-PT",
		"ja-JP",
		"zh-CN",
		"zh-TW",
		"ru-RU",
		"ko-KR",
	}
	screenWidths := []int{1920, 2560, 1366, 1440, 1536, 1600, 1680, 3840} // in pixels
	screenHeights := []int{1080, 1440, 768, 900, 864, 1024, 1050, 2160}   // in pixels
	colorDepths := []int{24, 32}                                          // in bits
	hardwareConcurrency := []int{2, 4, 6, 8, 12, 16}                      // number of logical cores
	deviceMemory := []int{2, 4, 8, 16}                                    // in GB
	maxTouchPoints := []int{0, 5, 10}                                     // number of touch points
	webglVendors := []string{
		"Google Inc. (Intel)",
		"Google Inc. (NVIDIA)",
		"Google Inc. (AMD)",
		"Google Inc. (Apple)",
		"Intel Inc.",
		"NVIDIA Corporation",
		"AMD",
		"Apple Inc.",
	}
	webglRenderers := []string{
		"ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1060 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"ANGLE (AMD, AMD Radeon RX 580 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		"Apple GPU",
		"Intel Iris Pro OpenGL Engine",
		"NVIDIA GeForce GTX 1060/PCIe/SSE2",
		"AMD Radeon RX 580",
	}
	userAgent := userAgents[rand.Intn(len(userAgents))]
	platform := platforms[rand.Intn(len(platforms))]
	language := languages[rand.Intn(len(languages))]
	languagesArray := []string{language, language[:2]}
	if rand.Float32() < 0.5 {
		languagesArray = append(languagesArray, "en")
	}
	screenWidth := screenWidths[rand.Intn(len(screenWidths))]
	screenHeight := screenHeights[rand.Intn(len(screenHeights))]
	colorDepth := colorDepths[rand.Intn(len(colorDepths))]
	hwConcurrency := hardwareConcurrency[rand.Intn(len(hardwareConcurrency))]
	devMemory := deviceMemory[rand.Intn(len(deviceMemory))]
	maxTouch := maxTouchPoints[rand.Intn(len(maxTouchPoints))]
	webglVendor := webglVendors[rand.Intn(len(webglVendors))]
	webglRenderer := webglRenderers[rand.Intn(len(webglRenderers))]
	return fingerprintConfig{
		UserAgent:           userAgent,
		Platform:            platform,
		Language:            language,
		Languages:           languagesArray,
		ScreenWidth:         screenWidth,
		ScreenHeight:        screenHeight,
		ColorDepth:          colorDepth,
		PixelDepth:          colorDepth,
		HardwareConcurrency: hwConcurrency,
		DeviceMemory:        devMemory,
		MaxTouchPoints:      maxTouch,
		WebGLVendor:         webglVendor,
		WebGLRenderer:       webglRenderer,
	}
}
