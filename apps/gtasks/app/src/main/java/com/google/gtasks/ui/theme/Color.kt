// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.google.gtasks.ui.theme

import androidx.compose.ui.graphics.Color

// Premium Neon Glassmorphic Palette based on Walmart Brand Colors
val BgMain = Color(0xFF041F41)            // Blue Ink (Deep Navy/Slate)
val BgCard = Color(0x73092E5C)            // Translucent Navy Card (rgba(9, 46, 92, 0.45))
val BgCardHover = Color(0x990F3D7A)       // Focused Card (rgba(15, 61, 122, 0.6))
val BgInput = Color(0x99031730)           // Text Fields (rgba(3, 23, 48, 0.6))

val BorderMuted = Color(0x0DFFFFFF)       // Muted Border (rgba(255, 255, 255, 0.05))
val BorderGlow = Color(0x260071CE)        // Walmart Blue Glow Border (rgba(0, 113, 206, 0.15))

val ColorPrimary = Color(0xFF0071CE)       // Walmart Blue
val ColorPrimaryGlow = Color(0x590071CE)   // Walmart Blue Glow (rgba(0, 113, 206, 0.35))
val ColorSecondary = Color(0xFFEB148D)     // Pink
val ColorSecondaryGlow = Color(0x40EB148D) // Pink Glow (rgba(235, 20, 141, 0.25))
val ColorAccent = Color(0xFFFFC220)        // Spark Yellow
val ColorAccentGlow = Color(0x40FFC220)     // Spark Yellow Glow (rgba(255, 194, 32, 0.25))

val ColorWarning = Color(0xFFF59E0B)       // Amber
val ColorCritical = Color(0xFFEF4444)      // Coral Red
val ColorCriticalGlow = Color(0x4DEF4444)  // Coral Red Glow (rgba(239, 68, 68, 0.3))

val TextPrimary = Color(0xFFF3F4F6)       // High-contrast Cool Grey
val TextSecondary = Color(0xFF9CA3AF)     // Medium-contrast Grey
val TextMuted = Color(0xFF6B7280)         // Muted Grey

// Light Mode Fallback Colors
val LightBgMain = Color(0xFFF0F4F8)        // Light Grey-Blue
val LightBgCard = Color(0xE6FFFFFF)        // Translucent White
val LightColorPrimary = Color(0xFF0071CE)  // Walmart Blue
val LightColorSecondary = Color(0xFFEB148D)// Pink
val LightColorAccent = Color(0xFFD9A000)     // Gold/Dark Yellow (accessible contrast)
val LightTextPrimary = Color(0xFF041F41)   // Blue Ink
val LightTextSecondary = Color(0xFF334E68) // Medium Navy
val LightTextMuted = Color(0xFF627D98)     // Muted Navy
val LightBorder = Color(0xFFD9E2EC)        // Light Border Grey
