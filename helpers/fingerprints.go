package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/chromedp/chromedp"
)

func RotateNavigatorFingerprints(ctx context.Context, intervalMs int) error {
	fingerprint := generateFingerprint()
	fingerprintJSON, err := json.Marshal(fingerprint)
	if err != nil {
		return fmt.Errorf("failed to marshal fingerprint: %v", err)
	}
	js := fmt.Sprintf(`(function() {
		const fingerprint = %s;
		// Override navigator.userAgent
		try {
			Object.defineProperty(navigator, 'userAgent', {
				get: function() { return fingerprint.userAgent; },
				configurable: true
			});
		} catch(e) {}
		// Override navigator.platform
		try {
			Object.defineProperty(navigator, 'platform', {
				get: function() { return fingerprint.platform; },
				configurable: true
			});
		} catch(e) {}
		// Override navigator.language
		try {
			Object.defineProperty(navigator, 'language', {
				get: function() { return fingerprint.language; },
				configurable: true
			});
		} catch(e) {}
		// Override navigator.languages
		try {
			Object.defineProperty(navigator, 'languages', {
				get: function() { return fingerprint.languages; },
				configurable: true
			});
		} catch(e) {}
		// Override navigator.hardwareConcurrency
		try {
			Object.defineProperty(navigator, 'hardwareConcurrency', {
				get: function() { return fingerprint.hardwareConcurrency; },
				configurable: true
			});
		} catch(e) {}
		// Override navigator.deviceMemory
		try {
			if ('deviceMemory' in navigator) {
				Object.defineProperty(navigator, 'deviceMemory', {
					get: function() { return fingerprint.deviceMemory; },
					configurable: true
				});
			}
		} catch(e) {}
		// Override navigator.maxTouchPoints
		try {
			Object.defineProperty(navigator, 'maxTouchPoints', {
				get: function() { return fingerprint.maxTouchPoints; },
				configurable: true
			});
		} catch(e) {}
		// Override screen properties
		try {
			Object.defineProperty(screen, 'width', {
				get: function() { return fingerprint.screenWidth; },
				configurable: true
			});
			Object.defineProperty(screen, 'height', {
				get: function() { return fingerprint.screenHeight; },
				configurable: true
			});
			Object.defineProperty(screen, 'availWidth', {
				get: function() { return fingerprint.screenWidth; },
				configurable: true
			});
			Object.defineProperty(screen, 'availHeight', {
				get: function() { return fingerprint.screenHeight - 40; },
				configurable: true
			});
			Object.defineProperty(screen, 'colorDepth', {
				get: function() { return fingerprint.colorDepth; },
				configurable: true
			});
			Object.defineProperty(screen, 'pixelDepth', {
				get: function() { return fingerprint.pixelDepth; },
				configurable: true
			});
		} catch(e) {}
		// Override WebGL vendor/renderer
		try {
			if (typeof WebGLRenderingContext !== 'undefined') {
				const getParameter = WebGLRenderingContext.prototype.getParameter;
				WebGLRenderingContext.prototype.getParameter = function(parameter) {
					if (parameter === 37445) {
						return fingerprint.webglVendor;
					}
					if (parameter === 37446) {
						return fingerprint.webglRenderer;
					}
					return getParameter.call(this, parameter);
				};
			}
		} catch(e) {}
		// Override Canvas fingerprinting with random noise
		try {
			const originalGetImageData = CanvasRenderingContext2D.prototype.getImageData;
			CanvasRenderingContext2D.prototype.getImageData = function(sx, sy, sw, sh) {
				const imageData = originalGetImageData.call(this, sx, sy, sw, sh);
				if (imageData && imageData.data) {
					for (let i = 0; i < imageData.data.length; i += 4) {
						const noise = (Math.random() - 0.5) * 2;
						imageData.data[i] = Math.max(0, Math.min(255, imageData.data[i] + noise));
					}
				}
				return imageData;
			};
			const originalToBlob = HTMLCanvasElement.prototype.toBlob;
			HTMLCanvasElement.prototype.toBlob = function(callback, type, quality) {
				const canvas = this;
				const ctx = canvas.getContext('2d');
				if (ctx) {
					try {
						const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
						if (imageData && imageData.data) {
							for (let i = 0; i < imageData.data.length; i += 4) {
								const noise = (Math.random() - 0.5) * 2;
								imageData.data[i] = Math.max(0, Math.min(255, imageData.data[i] + noise));
							}
							ctx.putImageData(imageData, 0, 0);
						}
					} catch(e) {}
				}
				return originalToBlob.call(this, callback, type, quality);
			};
			const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
			HTMLCanvasElement.prototype.toDataURL = function(type, quality) {
				const canvas = this;
				const ctx = canvas.getContext('2d');
				if (ctx) {
					try {
						const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
						if (imageData && imageData.data) {
							for (let i = 0; i < imageData.data.length; i += 4) {
								const noise = (Math.random() - 0.5) * 2;
								imageData.data[i] = Math.max(0, Math.min(255, imageData.data[i] + noise));
							}
							ctx.putImageData(imageData, 0, 0);
						}
					} catch(e) {}
				}
				return originalToDataURL.call(this, type, quality);
			};
		} catch(e) {}
	})();`, string(fingerprintJSON))
	var result interface{}
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &result)); err != nil {
		return fmt.Errorf("failed to inject fingerprint rotation script: %v", err)
	}
	return nil
}

func StartFingerprintRotation(ctx context.Context, intervalMs int) error {
	if intervalMs <= 0 {
		return nil
	}
	// Initial rotation
	if err := RotateNavigatorFingerprints(ctx, intervalMs); err != nil {
		return err
	}
	// Set up periodic rotation
	go func() {
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
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

func generateFingerprint() fingerprintConfig {
	rand.Seed(time.Now().UnixNano())
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
