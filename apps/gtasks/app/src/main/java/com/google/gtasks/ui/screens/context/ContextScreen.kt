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

package com.google.gtasks.ui.screens.context

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.Storefront
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.google.gtasks.data.model.Site
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic

@Composable
fun ContextScreen(
    onSiteSelected: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: ContextViewModel = viewModel(factory = ContextViewModel.Factory)
) {
    val uiState by viewModel.uiState.collectAsState()

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(
                Brush.radialGradient(
                    colors = listOf(
                        GTasksTheme.colors.colorAccent.copy(alpha = 0.12f),
                        GTasksTheme.colors.bgMain
                    ),
                    radius = 1200f
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp)
        ) {
            Spacer(modifier = Modifier.height(48.dp))

            // Title Header
            Text(
                text = "SELECT STORE",
                style = MaterialTheme.typography.displayMedium.copy(fontWeight = FontWeight.Black),
                color = GTasksTheme.colors.textPrimary
            )
            Text(
                text = "Establish your active storefront shift context",
                style = MaterialTheme.typography.bodyMedium,
                color = GTasksTheme.colors.textSecondary,
                modifier = Modifier.padding(top = 4.dp, bottom = 32.dp)
            )

            // Dynamic State Rendering
            when (val state = uiState) {
                is ContextUiState.Loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = GTasksTheme.colors.colorPrimary)
                    }
                }
                is ContextUiState.Success -> {
                    LazyColumn(
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                        modifier = Modifier.weight(1f)
                    ) {
                        items(state.sites) { site ->
                            SiteCard(
                                site = site,
                                onClick = {
                                    viewModel.selectSite(site)
                                    onSiteSelected()
                                }
                            )
                        }
                    }
                }
                is ContextUiState.Error -> {
                    Column(
                        modifier = Modifier.fillMaxSize(),
                        verticalArrangement = Arrangement.Center,
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = state.message,
                            color = GTasksTheme.colors.colorCritical,
                            style = MaterialTheme.typography.bodyLarge,
                            textAlign = TextAlign.Center
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(
                            onClick = { viewModel.loadSites() },
                            colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorPrimary)
                        ) {
                            Text("Retry")
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun SiteCard(
    site: Site,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .glassmorphic(shape = RoundedCornerShape(16.dp), elevation = 6.dp)
            .clickable(onClick = onClick)
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp)
        ) {
            // Storefront Icon Box
            Box(
                contentAlignment = Alignment.Center,
                modifier = Modifier
                    .size(48.dp)
                    .background(GTasksTheme.colors.colorPrimary.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
                    .glassmorphic(shape = RoundedCornerShape(12.dp), borderWidth = 1.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Storefront,
                    contentDescription = "Store Icon",
                    tint = GTasksTheme.colors.colorPrimary
                )
            }

            Spacer(modifier = Modifier.width(16.dp))

            // Site Details
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = site.name,
                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                    color = GTasksTheme.colors.textPrimary
                )
                Text(
                    text = "ID: ${site.id.take(8)}... | Org Context",
                    style = MaterialTheme.typography.bodySmall,
                    color = GTasksTheme.colors.textSecondary,
                    modifier = Modifier.padding(top = 2.dp)
                )
            }

            // Trailing Chevron
            Icon(
                imageVector = Icons.Default.ArrowForward,
                contentDescription = "Select Store Arrow",
                tint = GTasksTheme.colors.textSecondary,
                modifier = Modifier.size(20.dp)
            )
        }
    }
}
