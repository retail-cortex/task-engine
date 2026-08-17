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

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

/**
 * Applies a premium neon glassmorphic style to any composable.
 * Features a semi-transparent card background, drop shadow, and a thin white-indigo linear gradient border.
 */
@Composable
fun Modifier.glassmorphic(
    shape: Shape = RoundedCornerShape(16.dp),
    borderWidth: Dp = 1.dp,
    elevation: Dp = 6.dp
): Modifier {
    val themeColors = GTasksTheme.colors
    
    return this
        .shadow(
            elevation = elevation,
            shape = shape,
            clip = false,
            ambientColor = Color(0x20000000),
            spotColor = Color(0x40000000)
        )
        .background(
            color = themeColors.bgCard,
            shape = shape
        )
        .clip(shape)
        .border(
            width = borderWidth,
            brush = Brush.linearGradient(
                colors = listOf(
                    Color(0x20FFFFFF),           // 12% white highlight at top-left
                    Color(0x05FFFFFF),           // 2% white at center
                    themeColors.colorPrimary.copy(alpha = 0.15f) // 15% brand color glow at bottom-right
                )
            ),
            shape = shape
        )
}
