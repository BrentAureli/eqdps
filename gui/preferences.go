package main

import (
	"fmt"
	"runtime"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func settingToSlider(value, minimum, maximum float32) float32 {
	return (value - minimum) / (maximum - minimum)
}

func sliderToSetting(value, minimum, maximum float32) float32 {
	return minimum + value*(maximum-minimum)
}

func (s *shell) layoutPreferences(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return labelWeight(gtx, s.theme, "Preferences", unit.Sp(27), palette.text, text.Start, font.SemiBold)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset(0, unit.Dp(22)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return s.layoutPreferenceSlider(gtx, "Main window font scale", &s.mainScale, .75, 1.5, true)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset(0, unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return s.layoutPreferenceSlider(gtx, "DPS overlay font scale", &s.dpsScale, .75, 1.5, true)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return inset(0, unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return s.layoutPreferenceSlider(gtx, "DPS overlay opacity", &s.dpsOpacity, .35, 1, nativeOpacityAvailable())
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			message := "Opacity is stored, but this platform requires compositor configuration. See Help → Wayland overlay setup."
			if runtime.GOOS == "windows" {
				message = "Windows-native opacity support is the next implementation step."
			}
			return inset(0, unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(gtx, s.theme, message, unit.Sp(14), palette.muted, text.Start)
			})
		}),
	)
}

func (s *shell) layoutPreferenceSlider(gtx layout.Context, title string, state *widget.Float, minimum, maximum float32, enabled bool) layout.Dimensions {
	value := sliderToSetting(state.Value, minimum, maximum)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return label(gtx, s.theme, title, unit.Sp(17), palette.text, text.Start)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return labelWeight(gtx, s.theme, fmt.Sprintf("%d%%", int(value*100+.5)), unit.Sp(16), palette.accent, text.End, font.SemiBold)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = min(gtx.Constraints.Max.X, gtx.Dp(unit.Dp(520)))
			if !enabled {
				gtx = gtx.Disabled()
			}
			slider := material.Slider(s.theme, state)
			slider.Color = palette.accent
			dimensions := slider.Layout(gtx)
			s.applyPreferenceValues()
			return dimensions
		}),
	)
}

func (s *shell) applyPreferenceValues() {
	mainScale := sliderToSetting(s.mainScale.Value, .75, 1.5)
	dpsScale := sliderToSetting(s.dpsScale.Value, .75, 1.5)
	opacity := sliderToSetting(s.dpsOpacity.Value, .35, 1)
	if mainScale != s.settings.MainFontScale || dpsScale != s.settings.DPSFontScale || opacity != s.settings.DPSOpacity {
		s.settings.MainFontScale = mainScale
		s.settings.DPSFontScale = dpsScale
		s.settings.DPSOpacity = opacity
		s.theme.TextSize = unit.Sp(16 * mainScale)
		s.pushOverlay(s.fights)
		s.prefsDirty = true
	}
	if s.prefsDirty && !s.mainScale.Dragging() && !s.dpsScale.Dragging() && !s.dpsOpacity.Dragging() {
		s.prefsDirty = false
		if err := saveSettings(s.settings); err != nil {
			s.statusText = "Preferences could not be saved"
		}
	}
}

func nativeOpacityAvailable() bool {
	return false
}
