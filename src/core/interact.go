package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	chrome "logget/src/chrome"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

var eventKeyMap = map[string]input.Key{
	"Escape":      input.Escape,
	"Tab":         input.Tab,
	"Enter":       input.Enter,
	"Backspace":   input.Backspace,
	" ":           input.Space,
	"Control":     input.ControlLeft,
	"Meta":        input.MetaLeft,
	"Alt":         input.AltLeft,
	"Shift":       input.ShiftLeft,
	"ArrowLeft":   input.ArrowLeft,
	"ArrowRight":  input.ArrowRight,
	"ArrowUp":     input.ArrowUp,
	"ArrowDown":   input.ArrowDown,
	"Home":        input.Home,
	"End":         input.End,
	"PageUp":      input.PageUp,
	"PageDown":    input.PageDown,
	"Insert":      input.Insert,
	"Delete":      input.Delete,
	"CapsLock":    input.CapsLock,
	"F1":          input.F1,
	"F2":          input.F2,
	"F3":          input.F3,
	"F4":          input.F4,
	"F5":          input.F5,
	"F6":          input.F6,
	"F7":          input.F7,
	"F8":          input.F8,
	"F9":          input.F9,
	"F10":         input.F10,
	"F11":         input.F11,
	"F12":         input.F12,
	"ContextMenu": input.ContextMenu,
}

func RunInteractions(ctx *chrome.ChromeContext, specs []string) error {
	page := ctx.Page
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		parts := strings.SplitN(spec, ":", 3)
		if len(parts) < 2 {
			return fmt.Errorf("invalid interaction spec %q: expected action:value", spec)
		}
		action := strings.TrimSpace(strings.ToLower(parts[0]))
		rest := strings.TrimSpace(parts[1])
		var third string
		if len(parts) == 3 {
			third = strings.TrimSpace(parts[2])
		}
		switch action {
		case "wait":
			ms, err := strconv.Atoi(rest)
			if err != nil || ms < 0 {
				return fmt.Errorf("invalid interaction %q: wait requires non-negative ms", spec)
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			continue
		case "click":
			el, err := page.Element(rest)
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
		case "focus":
			el, err := page.Element(rest)
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			if err := el.ScrollIntoView(); err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			_, err = el.Evaluate(rod.Eval(`() => {
				var focusable = /^(A|BUTTON|INPUT|SELECT|TEXTAREA|SUMMARY|IFRAME|OBJECT|EMBED|DETAILS)$/i.test(this.tagName) || this.getAttribute('tabindex') != null;
				if (!focusable) this.tabIndex = -1;
				this.focus();
			}`).This(el.Object).ByUser())
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
		case "hover":
			el, err := page.Element(rest)
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			if err := el.Hover(); err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
		case "type":
			if len(parts) != 3 || rest == "" || third == "" {
				return fmt.Errorf("invalid interaction %q: type requires selector:text (e.g. type:#id:value)", spec)
			}
			selector := rest
			text := third
			el, err := page.Element(selector)
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			if err := el.Input(text); err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
		case "select":
			if rest == "" || (len(parts) == 3 && third == "") {
				return fmt.Errorf("invalid interaction %q: select requires selector and value", spec)
			}
			selector := rest
			value := third
			if len(parts) == 2 {
				return fmt.Errorf("invalid interaction %q: select requires selector:value", spec)
			}
			el, err := page.Element(selector)
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			if err := el.Select([]string{value}, true, rod.SelectorTypeText); err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
		case "key":
			var selector, keySpec string
			if len(parts) == 3 {
				selector = rest
				keySpec = third
			} else {
				keySpec = rest
			}
			if keySpec == "" {
				return fmt.Errorf("invalid interaction %q: key requires key or selector:key", spec)
			}
			if selector != "" {
				el, err := page.Element(selector)
				if err != nil {
					return fmt.Errorf("interaction %q: %w", spec, err)
				}
				if err := el.Focus(); err != nil {
					return fmt.Errorf("interaction %q: %w", spec, err)
				}
			}
			keys, err := parseKeySpec(keySpec)
			if err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
			if err := pressKeys(page, keys); err != nil {
				return fmt.Errorf("interaction %q: %w", spec, err)
			}
		default:
			return fmt.Errorf("unknown action %q in interaction %q", action, spec)
		}
	}
	return nil
}

func parseKeySpec(keySpec string) ([]input.Key, error) {
	parts := strings.Split(keySpec, "+")
	keys := make([]input.Key, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		k, err := eventKeyToRodKey(p)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("empty key spec")
	}
	return keys, nil
}

func eventKeyToRodKey(name string) (input.Key, error) {
	if name == "" {
		return 0, fmt.Errorf("empty key name")
	}
	if k, ok := eventKeyMap[name]; ok {
		return k, nil
	}
	if len(name) == 1 {
		return input.Key(name[0]), nil
	}
	return input.AddKey(name, "", name, 0, 0), nil
}

func pressKeys(page *rod.Page, keys []input.Key) error {
	if len(keys) == 1 {
		return page.Keyboard.Press(keys[0])
	}
	ka := page.KeyActions()
	for i := 0; i < len(keys)-1; i++ {
		ka.Press(keys[i])
	}
	ka.Type(keys[len(keys)-1])
	return ka.Do()
}
