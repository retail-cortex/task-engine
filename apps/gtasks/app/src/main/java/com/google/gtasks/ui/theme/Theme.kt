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

import android.app.Activity
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.SideEffect
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

// 1. Custom Design Tokens Holder (extends standard Material3 color scheme)
data class GTasksColors(
    val bgMain: Color,
    val bgCard: Color,
    val bgCardHover: Color,
    val bgInput: Color,
    val borderMuted: Color,
    val borderGlow: Color,
    val colorPrimary: Color,
    val colorSecondary: Color,
    val colorAccent: Color,
    val colorCritical: Color,
    val colorWarning: Color,
    val textPrimary: Color,
    val textSecondary: Color,
    val textMuted: Color
)

// Dark Theme Custom Tokens
private val DarkGTasksColors = GTasksColors(
    bgMain = BgMain,
    bgCard = BgCard,
    bgCardHover = BgCardHover,
    bgInput = BgInput,
    borderMuted = BorderMuted,
    borderGlow = BorderGlow,
    colorPrimary = ColorPrimary,
    colorSecondary = ColorSecondary,
    colorAccent = ColorAccent,
    colorCritical = ColorCritical,
    colorWarning = ColorWarning,
    textPrimary = TextPrimary,
    textSecondary = TextSecondary,
    textMuted = TextMuted
)

// Light Theme Custom Tokens
private val LightGTasksColors = GTasksColors(
    bgMain = LightBgMain,
    bgCard = LightBgCard,
    bgCardHover = LightBgCard, // same for light
    bgInput = LightBgCard,
    borderMuted = LightBorder,
    borderGlow = LightBorder,
    colorPrimary = LightColorPrimary,
    colorSecondary = LightColorSecondary,
    colorAccent = LightColorAccent,
    colorCritical = ColorCritical, // same critical
    colorWarning = ColorWarning,
    textPrimary = LightTextPrimary,
    textSecondary = LightTextSecondary,
    textMuted = LightTextMuted
)

private val LocalGTasksColors = staticCompositionLocalOf { DarkGTasksColors }

// Global accessor object
object GTasksTheme {
    val colors: GTasksColors
        @Composable
        @ReadOnlyComposable
        get() = LocalGTasksColors.current
}

// 2. Standard Material3 Color Schemes
private val DarkColorScheme = darkColorScheme(
    primary = ColorPrimary,
    secondary = ColorSecondary,
    tertiary = ColorAccent,
    background = BgMain,
    surface = BgCard,
    onPrimary = TextPrimary,
    onSecondary = TextPrimary,
    onBackground = TextPrimary,
    onSurface = TextPrimary,
    error = ColorCritical
)

private val LightColorScheme = lightColorScheme(
    primary = LightColorPrimary,
    secondary = LightColorSecondary,
    tertiary = LightColorAccent,
    background = LightBgMain,
    surface = LightBgCard,
    onPrimary = Color.White,
    onSecondary = Color.White,
    onBackground = LightTextPrimary,
    onSurface = LightTextPrimary,
    error = ColorCritical
)

@Composable
fun GTasksTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) DarkColorScheme else LightColorScheme
    val customColors = if (darkTheme) DarkGTasksColors else LightGTasksColors

    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            window.statusBarColor = Color.Transparent.toArgb()
            window.navigationBarColor = Color.Transparent.toArgb()
            
            // Adjust status bar icon colors based on theme
            val windowInsetsController = WindowCompat.getInsetsController(window, view)
            windowInsetsController.isAppearanceLightStatusBars = !darkTheme
            windowInsetsController.isAppearanceLightNavigationBars = !darkTheme
        }
    }

    CompositionLocalProvider(
        LocalGTasksColors provides customColors
    ) {
        MaterialTheme(
            colorScheme = colorScheme,
            typography = Typography,
            content = content
        )
    }
}
