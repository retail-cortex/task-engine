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

package com.google.gtasks.ui.screens.detail

import androidx.compose.animation.core.*
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.nativeCanvas
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic

// Layout Types
enum class StoreLayout { LINEAR, BOUTIQUE, RACETRACK }

/**
 * Maps a Site ID to its corresponding store blueprint layout.
 */
fun getStoreLayout(siteId: String): StoreLayout {
    return when (siteId) {
        "44444444-4444-4444-4444-444444440001" -> StoreLayout.BOUTIQUE // San Francisco
        "44444444-4444-4444-4444-444444440002" -> StoreLayout.RACETRACK // Los Angeles
        else -> StoreLayout.LINEAR // Seattle/Dallas & default fallback
    }
}

/**
 * Resolves the location name from a task title.
 */
fun resolveLocationName(taskName: String): String {
    val name = taskName.lowercase();
    return when {
        name.includes("aisle 7") -> "Aisle 7"
        name.includes("aisle 8") -> "Aisle 8"
        name.includes("aisle 9") -> "Aisle 9"
        name.includes("aisle 10") -> "Aisle 10"
        name.includes("produce") || name.includes("greens") -> "Produce & Greens"
        name.includes("deli") || name.includes("hot food") -> "Deli Depot"
        name.includes("bakery") -> "Bakery Ovens"
        name.includes("cooler") -> "Walk-in Cooler"
        name.includes("pharmacy") -> "Pharmacy"
        name.includes("vault") -> "Secure Vault"
        name.includes("dock") || name.includes("loading") -> "Loading Dock"
        name.includes("stock") || name.includes("cage") -> "Stock Cage"
        name.includes("checkout") || name.includes("register") -> "Checkout Counters"
        name.includes("lounge") -> "Lounge Area"
        name.includes("atrium") -> "Atrium Experience"
        else -> "Store Floor"
    }
}

private fun String.includes(substring: String): Boolean = this.contains(substring, ignoreCase = true)

/**
 * Translates location name and store layout to the SVG blueprint coordinate grid (0..200 x 0..150).
 */
fun getFocalCoordinates(locationName: String, layout: StoreLayout): Offset? {
    val name = locationName.lowercase();
    return when (layout) {
        StoreLayout.BOUTIQUE -> when {
            name.includes("loading") -> Offset(16f, 25f)
            name.includes("stock") || name.includes("cage") -> Offset(16f, 105f)
            name.includes("produce") || name.includes("greens") || name.includes("wall") -> Offset(65f, 10f)
            name.includes("vault") -> Offset(175f, 23f)
            name.includes("showcase a") -> Offset(100f, 75f)
            name.includes("showcase b") -> Offset(60f, 75f)
            name.includes("showcase c") -> Offset(140f, 75f)
            name.includes("lounge") || name.includes("coffee") -> Offset(170f, 126f)
            name.includes("checkout") || name.includes("register") || name.includes("counter") -> Offset(105f, 126f)
            else -> Offset(100f, 75f) // Center fallback
        }
        StoreLayout.RACETRACK -> when {
            name.includes("atrium") -> Offset(100f, 72f)
            name.includes("intake") || name.includes("dock") -> Offset(180f, 17f)
            name.includes("staging") -> Offset(180f, 62f)
            name.includes("vault") -> Offset(25f, 127f)
            name.includes("register") || name.includes("gallery") || name.includes("checkout") -> Offset(140f, 129f)
            name.includes("canopy") || name.includes("fresh") || name.includes("produce") -> Offset(32.5f, 12f)
            name.includes("perishables") || name.includes("market") -> Offset(110f, 12f)
            name.includes("aisle a") || name.includes("aisle 1") -> Offset(25f, 65f)
            name.includes("aisle b") || name.includes("aisle 2") -> Offset(145f, 65f)
            else -> Offset(100f, 75f)
        }
        StoreLayout.LINEAR -> when {
            name.includes("dock a") -> Offset(16f, 12f)
            name.includes("dock b") -> Offset(16f, 30f)
            name.includes("dock") || name.includes("loading") -> Offset(16f, 21f)
            name.includes("stock") || name.includes("backroom") -> Offset(16f, 98f)
            name.includes("vault") -> Offset(180f, 22f)
            name.includes("produce") || name.includes("greens") -> Offset(95f, 12f)
            name.includes("deli") || name.includes("hot food") -> Offset(95f, 34f)
            name.includes("bakery") -> Offset(180f, 55f)
            name.includes("cooler") -> Offset(180f, 108f)
            name.includes("pharmacy") -> Offset(146f, 82.5f)
            name.includes("aisle 7") -> Offset(46f, 82.5f)
            name.includes("aisle 8") -> Offset(62f, 82.5f)
            name.includes("aisle 9") -> Offset(78f, 82.5f)
            name.includes("aisle 10") -> Offset(94f, 82.5f)
            name.includes("register 1") || name.includes("checkout 1") -> Offset(114f, 62.5f)
            name.includes("register 2") || name.includes("checkout 2") -> Offset(114f, 94.5f)
            name.includes("register 3") || name.includes("checkout 3") -> Offset(130f, 62.5f)
            name.includes("register 4") || name.includes("checkout 4") -> Offset(130f, 94.5f)
            name.includes("register") || name.includes("checkout") -> Offset(122f, 78.5f)
            else -> Offset(100f, 75f)
        }
    }
}

@Composable
fun StoreMapDialog(
    siteId: String,
    siteName: String,
    locationName: String,
    onDismiss: () -> Unit
) {
    val layout = getStoreLayout(siteId)
    val focalPoint = getFocalCoordinates(locationName, layout)

    // Setup pulsing beacon animations
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val beaconRadius by infiniteTransition.animateFloat(
        initialValue = 4f,
        targetValue = 14f,
        animationSpec = infiniteRepeatable(
            animation = tween(1400, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "radius"
    )
    val beaconOpacity by infiniteTransition.animateFloat(
        initialValue = 0.8f,
        targetValue = 0f,
        animationSpec = infiniteRepeatable(
            animation = tween(1400, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "opacity"
    )

    Dialog(onDismissRequest = onDismiss) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .glassmorphic(shape = RoundedCornerShape(24.dp), elevation = 12.dp)
                .padding(20.dp)
        ) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                // Header
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column {
                        Text(
                            text = "Store Digital Twin",
                            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                            color = GTasksTheme.colors.textPrimary
                        )
                        Text(
                            text = "$siteName • ${layout.name} Layout",
                            style = MaterialTheme.typography.labelSmall,
                            color = GTasksTheme.colors.textSecondary
                        )
                    }
                    IconButton(
                        onClick = onDismiss,
                        colors = IconButtonDefaults.iconButtonColors(contentColor = GTasksTheme.colors.textSecondary)
                    ) {
                        Icon(Icons.Default.Close, contentDescription = "Close Map")
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                // The responsive vector blueprint drawing Canvas
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(4f / 3f) // Standard 4:3 aspect ratio matching the 200x150 blueprint
                        .background(GTasksTheme.colors.bgInput, RoundedCornerShape(12.dp))
                        .border(1.dp, GTasksTheme.colors.borderMuted, RoundedCornerShape(12.dp))
                        .padding(12.dp)
                ) {
                    val primaryColor = GTasksTheme.colors.colorPrimary
                    Canvas(modifier = Modifier.fillMaxSize()) {
                        val scaleX = size.width / 200f
                        val scaleY = size.height / 150f

                        // Draw outer dashed bounds
                        drawRoundRect(
                            color = primaryColor.copy(alpha = 0.2f),
                            topLeft = Offset(2f * scaleX, 2f * scaleY),
                            size = Size(196f * scaleX, 146f * scaleY),
                            cornerRadius = CornerRadius(3f * scaleX),
                            style = Stroke(
                                width = 1.dp.toPx(),
                                pathEffect = PathEffect.dashPathEffect(floatArrayOf(10f, 5f), 0f)
                            )
                        )

                        // Draw specific store blueprints
                        when (layout) {
                            StoreLayout.BOUTIQUE -> drawBoutiqueLayout(scaleX, scaleY)
                            StoreLayout.RACETRACK -> drawRacetrackLayout(scaleX, scaleY)
                            StoreLayout.LINEAR -> drawLinearLayout(scaleX, scaleY)
                        }

                        // Draw Pulsing Beacon if matched
                        focalPoint?.let { point ->
                            val fx = point.x * scaleX
                            val fy = point.y * scaleY

                            // 1. Outer Pulsing Glow Circle
                            drawCircle(
                                color = Color(0xFFEF4444),
                                radius = beaconRadius * scaleX,
                                center = Offset(fx, fy),
                                alpha = beaconOpacity
                            )
                            // 2. Inner Solid Core Circle
                            drawCircle(
                                color = Color(0xFFEF4444),
                                radius = 3f * scaleX,
                                center = Offset(fx, fy)
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                // HUD Metadata Panel
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(GTasksTheme.colors.bgInput, RoundedCornerShape(12.dp))
                        .border(1.dp, GTasksTheme.colors.borderMuted, RoundedCornerShape(12.dp))
                        .padding(12.dp)
                ) {
                    Column {
                        Text(
                            text = "Focal Targeting System",
                            style = MaterialTheme.typography.labelMedium.copy(fontWeight = FontWeight.Bold),
                            color = GTasksTheme.colors.textPrimary
                        )
                        Spacer(modifier = Modifier.height(6.dp))
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .size(8.dp)
                                    .background(Color(0xFFEF4444), RoundedCornerShape(50))
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = "Zone Locked: $locationName",
                                style = MaterialTheme.typography.bodySmall,
                                color = GTasksTheme.colors.textSecondary
                            )
                        }
                    }
                }
            }
        }
    }
}

// -------------------------------------------------------------------------------------
// Vector Blueprint Draw Handlers (Replicating Web SVGs)
// -------------------------------------------------------------------------------------

private fun DrawScope.drawBoutiqueLayout(scaleX: Float, scaleY: Float) {
    // 1. Loading Bay
    drawFixture(5f, 5f, 22f, 40f, "LOADING", scaleX, scaleY)
    // 2. Stock Cage
    drawFixture(5f, 65f, 22f, 78f, "STOCK CAGE", scaleX, scaleY)
    // 3. Organic Greens
    drawFixture(35f, 5f, 60f, 10f, "GREENS WALL", scaleX, scaleY)
    // 4. Secure Vault
    drawFixture(155f, 5f, 40f, 35f, "VAULT", scaleX, scaleY)
    // 5. Showcase A, B, C
    drawCircleFixture(100f, 75f, 20f, "SHOWCASE A", scaleX, scaleY)
    drawCircleFixture(60f, 75f, 12f, "SHOWCASE B", scaleX, scaleY)
    drawCircleFixture(140f, 75f, 12f, "SHOWCASE C", scaleX, scaleY)
    // 6. Lounge
    drawFixture(145f, 110f, 50f, 33f, "LOUNGE", scaleX, scaleY)
    // 7. Checkout
    drawFixture(90f, 120f, 30f, 12f, "CHECKOUT", scaleX, scaleY)
}

private fun DrawScope.drawRacetrackLayout(scaleX: Float, scaleY: Float) {
    // 1. Atrium
    drawFixture(50f, 40f, 100f, 65f, "ATRIUM CENTER", scaleX, scaleY, isDashed = true)
    // 2. Intake Bay
    drawFixture(165f, 5f, 30f, 25f, "INTAKE", scaleX, scaleY)
    // 3. Staging
    drawFixture(165f, 35f, 30f, 55f, "STAGING", scaleX, scaleY)
    // 4. Vault
    drawFixture(5f, 110f, 40f, 35f, "VAULT", scaleX, scaleY)
    // 5. Register
    drawFixture(120f, 115f, 40f, 28f, "REGISTER GALLERY", scaleX, scaleY)
    // 6. Canopy
    drawFixture(5f, 5f, 55f, 15f, "CANOPY", scaleX, scaleY)
    // 7. Perishables
    drawFixture(65f, 5f, 90f, 15f, "MARKET", scaleX, scaleY)
    // 8. Aisles
    drawFixture(20f, 30f, 10f, 70f, "AISLE A", scaleX, scaleY)
    drawFixture(140f, 30f, 10f, 70f, "AISLE B", scaleX, scaleY)
}

private fun DrawScope.drawLinearLayout(scaleX: Float, scaleY: Float) {
    // 1. Docks
    drawFixture(5f, 5f, 22f, 15f, "DOCK A", scaleX, scaleY)
    drawFixture(5f, 23f, 22f, 15f, "DOCK B", scaleX, scaleY)
    // 2. Stock Room
    drawFixture(5f, 50f, 22f, 95f, "STOCK ROOM", scaleX, scaleY)
    // 3. Vault
    drawFixture(165f, 5f, 30f, 35f, "VAULT", scaleX, scaleY)
    // 4. Produce
    drawFixture(40f, 5f, 110f, 15f, "PRODUCE WET WALL", scaleX, scaleY)
    // 5. Deli
    drawFixture(40f, 28f, 110f, 12f, "DELI DEPOT", scaleX, scaleY)
    // 6. Bakery
    drawFixture(165f, 48f, 30f, 15f, "BAKERY", scaleX, scaleY)
    // 7. Aisles
    drawFixture(42f, 50f, 8f, 65f, "A7A", scaleX, scaleY)
    drawFixture(58f, 50f, 8f, 65f, "A8A", scaleX, scaleY)
    drawFixture(74f, 50f, 8f, 65f, "A9A", scaleX, scaleY)
    drawFixture(90f, 50f, 8f, 65f, "A10A", scaleX, scaleY)
    drawFixture(42f, 122f, 8f, 23f, "A7B", scaleX, scaleY)
    drawFixture(58f, 122f, 8f, 23f, "A8B", scaleX, scaleY)
    drawFixture(74f, 122f, 8f, 23f, "A9B", scaleX, scaleY)
    drawFixture(90f, 122f, 8f, 23f, "A10B", scaleX, scaleY)
    // 8. Registers
    drawFixture(110f, 50f, 8f, 25f, "R1", scaleX, scaleY)
    drawFixture(110f, 82f, 8f, 25f, "R2", scaleX, scaleY)
    drawFixture(126f, 50f, 8f, 25f, "R3", scaleX, scaleY)
    drawFixture(126f, 82f, 8f, 25f, "R4", scaleX, scaleY)
    // 9. Pharmacy & Cooler
    drawFixture(142f, 50f, 8f, 65f, "PHARMACY", scaleX, scaleY)
    drawFixture(165f, 70f, 30f, 75f, "COOLER", scaleX, scaleY)
}

// -------------------------------------------------------------------------------------
// Drawing Helpers
// -------------------------------------------------------------------------------------

private fun DrawScope.drawFixture(
    x: Float, y: Float, w: Float, h: Float,
    label: String,
    scaleX: Float, scaleY: Float,
    isDashed: Boolean = false,
    fixtureColor: Color = Color(0xFF0071CE)
) {
    val tx = x * scaleX
    val ty = y * scaleY
    val tw = w * scaleX
    val th = h * scaleY

    // 1. Draw Semi-translucent Fill
    drawRoundRect(
        color = fixtureColor.copy(alpha = 0.07f),
        topLeft = Offset(tx, ty),
        size = Size(tw, th),
        cornerRadius = CornerRadius(2f * scaleX)
    )

    // 2. Draw Sleek Neon Border
    drawRoundRect(
        color = fixtureColor.copy(alpha = 0.33f),
        topLeft = Offset(tx, ty),
        size = Size(tw, th),
        cornerRadius = CornerRadius(2f * scaleX),
        style = Stroke(
            width = 0.8f.dp.toPx(),
            pathEffect = if (isDashed) PathEffect.dashPathEffect(floatArrayOf(5f, 3f), 0f) else null
        )
    )
}

private fun DrawScope.drawCircleFixture(
    cx: Float, cy: Float, r: Float,
    label: String,
    scaleX: Float, scaleY: Float,
    fixtureColor: Color = Color(0xFF0071CE)
) {
    val tcx = cx * scaleX
    val tcy = cy * scaleY
    val tr = r * scaleX

    // Fill
    drawCircle(
        color = fixtureColor.copy(alpha = 0.07f),
        radius = tr,
        center = Offset(tcx, tcy)
    )

    // Border
    drawCircle(
        color = fixtureColor.copy(alpha = 0.33f),
        radius = tr,
        center = Offset(tcx, tcy),
        style = Stroke(width = 0.8f.dp.toPx())
    )
}
