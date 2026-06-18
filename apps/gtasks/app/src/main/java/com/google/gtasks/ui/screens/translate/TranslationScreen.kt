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

package com.google.gtasks.ui.screens.translate

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.core.*
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TranslationScreen(
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: TranslationViewModel = viewModel(factory = TranslationViewModel.Factory)
) {
    val context = LocalContext.current
    val uiState by viewModel.uiState.collectAsState()
    val isRecording by viewModel.isRecording.collectAsState()
    val activeMode by viewModel.activeRecordingMode.collectAsState()
    val conversation by viewModel.conversation.collectAsState()
    val voices by viewModel.voices.collectAsState()
    val profile by viewModel.profile.collectAsState()
    val errorMessage by viewModel.errorMessage.collectAsState()

    var selectedTabIndex by remember { mutableIntStateOf(0) }
    var targetLanguage by remember { mutableStateOf("es-ES") } // Default to Spanish
    var customerGender by remember { mutableStateOf("MALE") }
    
    // Dropdown selectors states
    var langDropdownExpanded by remember { mutableStateOf(false) }
    var voiceDropdownExpanded by remember { mutableStateOf(false) }

    val coroutineScope = rememberCoroutineScope()
    val listState = rememberLazyListState()

    // Request Microphone Permission
    var hasMicPermission by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(context, Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED
        )
    }

    val launcher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        hasMicPermission = isGranted
    }

    LaunchedEffect(Unit) {
        if (!hasMicPermission) {
            launcher.launch(Manifest.permission.RECORD_AUDIO)
        }
    }

    // Auto-scroll to the latest translation bubble in the history
    LaunchedEffect(conversation.size) {
        if (conversation.isNotEmpty()) {
            coroutineScope.launch {
                listState.animateScrollToItem(conversation.size - 1)
            }
        }
    }

    Scaffold(
        topBar = {
            // Premium Gradient Top Header
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(
                        brush = Brush.horizontalGradient(
                            colors = listOf(
                                GTasksTheme.colors.colorPrimary,
                                GTasksTheme.colors.colorSecondary
                            )
                        )
                    )
                    .statusBarsPadding()
                    .padding(horizontal = 4.dp, vertical = 8.dp)
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    IconButton(
                        onClick = onBackClick,
                        colors = IconButtonDefaults.iconButtonColors(contentColor = Color.White)
                    ) {
                        Icon(
                            imageVector = Icons.Default.ArrowBack,
                            contentDescription = "Go Back"
                        )
                    }
                    Text(
                        text = "REAL-TIME TRANSLATOR",
                        style = MaterialTheme.typography.titleMedium.copy(
                            fontWeight = FontWeight.Black,
                            letterSpacing = 1.5.sp
                        ),
                        color = Color.White,
                        modifier = Modifier.padding(start = 8.dp)
                    )
                }
            }
        }
    ) { innerPadding ->
        Column(
            modifier = modifier
                .fillMaxSize()
                .padding(innerPadding)
                .background(GTasksTheme.colors.bgMain)
        ) {
            // Error Banner
            errorMessage?.let { msg ->
                Card(
                    colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.colorCritical.copy(alpha = 0.1f)),
                    border = BorderStroke(1.dp, GTasksTheme.colors.colorCritical.copy(alpha = 0.4f)),
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Default.Cancel,
                            tint = GTasksTheme.colors.colorCritical,
                            contentDescription = "Error"
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = msg,
                            color = GTasksTheme.colors.colorCritical,
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                }
            }

            // Navigation Tab Row (Translate vs Settings)
            TabRow(
                selectedTabIndex = selectedTabIndex,
                containerColor = Color.Transparent,
                contentColor = GTasksTheme.colors.textPrimary,
                indicator = { tabPositions ->
                    TabRowDefaults.SecondaryIndicator(
                        Modifier.tabIndicatorOffset(tabPositions[selectedTabIndex]),
                        color = GTasksTheme.colors.colorPrimary
                    )
                }
            ) {
                Tab(
                    selected = selectedTabIndex == 0,
                    onClick = { selectedTabIndex = 0 },
                    text = {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(imageVector = Icons.Default.Translate, contentDescription = null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(text = "Translate", fontWeight = FontWeight.Bold)
                        }
                    }
                )
                Tab(
                    selected = selectedTabIndex == 1,
                    onClick = { selectedTabIndex = 1 },
                    text = {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(imageVector = Icons.Default.Settings, contentDescription = null)
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(text = "Voice Settings", fontWeight = FontWeight.Bold)
                        }
                    }
                )
            }

            if (selectedTabIndex == 0) {
                // TAB 1: TRANSLATION CONSOLE
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .padding(16.dp)
                ) {
                    // 1. Language Selector Card
                    Card(
                        colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                        border = BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 16.dp)
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            // Left: Associate Language (Read only)
                            Column(horizontalAlignment = Alignment.Start) {
                                Text(
                                    text = "ASSOCIATE",
                                    style = MaterialTheme.typography.labelSmall,
                                    color = GTasksTheme.colors.textMuted
                                )
                                Text(
                                    text = profile?.preferredLanguageId?.let { "English (US)" } ?: "English",
                                    style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                    color = GTasksTheme.colors.textPrimary
                                )
                            }

                            // Swap Arrow Icon
                            Icon(
                                imageVector = Icons.Default.SwapHoriz,
                                contentDescription = "Swap Languages",
                                tint = GTasksTheme.colors.colorPrimary,
                                modifier = Modifier.size(28.dp)
                            )

                            // Right: Customer Language (Dropdown Selector)
                            Column(horizontalAlignment = Alignment.End) {
                                Text(
                                    text = "CUSTOMER",
                                    style = MaterialTheme.typography.labelSmall,
                                    color = GTasksTheme.colors.textMuted
                                )
                                Box {
                                    val languages = remember {
                                        listOf(
                                            "ar-XA" to "Arabic",
                                            "zh-CN" to "Chinese (Simplified)",
                                            "nl-NL" to "Dutch",
                                            "en-US" to "English",
                                            "fr-FR" to "French",
                                            "de-DE" to "German",
                                            "hi-IN" to "Hindi",
                                            "it-IT" to "Italian",
                                            "ja-JP" to "Japanese",
                                            "ko-KR" to "Korean",
                                            "pl-PL" to "Polish",
                                            "pt-BR" to "Portuguese (Brazil)",
                                            "ru-RU" to "Russian",
                                            "es-ES" to "Spanish",
                                            "sv-SE" to "Swedish",
                                            "tr-TR" to "Turkish",
                                            "vi-VN" to "Vietnamese"
                                        )
                                    }

                                    Row(
                                        verticalAlignment = Alignment.CenterVertically,
                                        modifier = Modifier.clickable { langDropdownExpanded = true }
                                    ) {
                                        Text(
                                            text = languages.firstOrNull { it.first == targetLanguage }?.second ?: "Spanish",
                                            style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                            color = GTasksTheme.colors.colorPrimary
                                        )
                                        Icon(
                                            imageVector = Icons.Default.ArrowDropDown,
                                            tint = GTasksTheme.colors.colorPrimary,
                                            contentDescription = null
                                        )
                                    }
                                    DropdownMenu(
                                        expanded = langDropdownExpanded,
                                        onDismissRequest = { langDropdownExpanded = false },
                                        modifier = Modifier
                                            .background(GTasksTheme.colors.bgMain)
                                            .border(1.dp, GTasksTheme.colors.borderMuted)
                                    ) {
                                        languages.forEach { (code, name) ->
                                            DropdownMenuItem(
                                                text = { Text(name) },
                                                onClick = {
                                                    targetLanguage = code
                                                    langDropdownExpanded = false
                                                }
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }

                    // 2. Customer Gender Selector Card (For listen playback TTS style)
                    Card(
                        colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                        border = BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 16.dp)
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 16.dp, vertical = 12.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Text(
                                text = "Customer Voice Style:",
                                style = MaterialTheme.typography.bodyMedium,
                                color = GTasksTheme.colors.textSecondary
                            )
                            Row {
                                FilterChip(
                                    selected = customerGender == "MALE",
                                    onClick = { customerGender = "MALE" },
                                    label = { Text("Male") },
                                    modifier = Modifier.padding(end = 8.dp)
                                )
                                FilterChip(
                                    selected = customerGender == "FEMALE",
                                    onClick = { customerGender = "FEMALE" },
                                    label = { Text("Female") }
                                )
                            }
                        }
                    }

                    // 3. Chat History Stream Panel
                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth()
                            .border(1.dp, GTasksTheme.colors.borderMuted, RoundedCornerShape(12.dp))
                            .background(GTasksTheme.colors.bgCard, RoundedCornerShape(12.dp))
                            .padding(12.dp)
                    ) {
                        if (conversation.isEmpty()) {
                            Column(
                                modifier = Modifier.fillMaxSize(),
                                verticalArrangement = Arrangement.Center,
                                horizontalAlignment = Alignment.CenterHorizontally
                            ) {
                                Icon(
                                    imageVector = Icons.Default.MicNone,
                                    contentDescription = null,
                                    tint = GTasksTheme.colors.textMuted,
                                    modifier = Modifier.size(48.dp)
                                )
                                Spacer(modifier = Modifier.height(12.dp))
                                Text(
                                    text = "Ready to translate!\nTap and hold a microphone button below to talk.",
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = GTasksTheme.colors.textMuted,
                                    textAlign = TextAlign.Center
                                )
                            }
                        } else {
                            LazyColumn(
                                state = listState,
                                modifier = Modifier.fillMaxSize(),
                                verticalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                items(conversation) { item ->
                                    val isAssociate = item.speaker == Speaker.ASSOCIATE
                                    Row(
                                        modifier = Modifier.fillMaxWidth(),
                                        horizontalArrangement = if (isAssociate) Arrangement.End else Arrangement.Start
                                    ) {
                                        Card(
                                            shape = RoundedCornerShape(
                                                topStart = 16.dp,
                                                topEnd = 16.dp,
                                                bottomStart = if (isAssociate) 16.dp else 2.dp,
                                                bottomEnd = if (isAssociate) 2.dp else 16.dp
                                            ),
                                            colors = CardDefaults.cardColors(
                                                containerColor = if (isAssociate) {
                                                    GTasksTheme.colors.colorPrimary
                                                } else {
                                                    GTasksTheme.colors.borderMuted.copy(alpha = 0.3f)
                                                }
                                            ),
                                            border = if (isAssociate) null else BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                                            modifier = Modifier.widthIn(max = 280.dp)
                                        ) {
                                            Column(modifier = Modifier.padding(12.dp)) {
                                                Row(
                                                    verticalAlignment = Alignment.CenterVertically,
                                                    horizontalArrangement = Arrangement.SpaceBetween,
                                                    modifier = Modifier.fillMaxWidth()
                                                ) {
                                                    Text(
                                                        text = if (isAssociate) "Associate (You)" else "Customer",
                                                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                                                        color = if (isAssociate) Color.White.copy(alpha = 0.8f) else GTasksTheme.colors.textSecondary
                                                    )
                                                    
                                                    // Replay Audio Trigger Button!
                                                    if (item.audioBytes != null) {
                                                        Icon(
                                                            imageVector = Icons.Default.VolumeUp,
                                                            contentDescription = "Replay Translation",
                                                            tint = if (isAssociate) Color.White else GTasksTheme.colors.colorPrimary,
                                                            modifier = Modifier
                                                                .size(18.dp)
                                                                .clickable { viewModel.playAudioBytes(item.audioBytes) }
                                                        )
                                                    }
                                                }
                                                Spacer(modifier = Modifier.height(4.dp))
                                                Text(
                                                    text = item.translatedText,
                                                    style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                                    color = if (isAssociate) Color.White else GTasksTheme.colors.textPrimary
                                                )
                                            }
                                        }
                                    }
                                }
                            }
                        }

                        // Translating Overlay State Spinner
                        if (uiState is TranslationUiState.Translating) {
                            Box(
                                modifier = Modifier
                                    .fillMaxSize()
                                    .background(Color.Black.copy(alpha = 0.3f), RoundedCornerShape(12.dp)),
                                contentAlignment = Alignment.Center
                            ) {
                                Card(
                                    colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                                    modifier = Modifier.padding(24.dp)
                                ) {
                                    Row(
                                        modifier = Modifier.padding(16.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        CircularProgressIndicator(modifier = Modifier.size(24.dp))
                                        Spacer(modifier = Modifier.width(12.dp))
                                        Text(text = "Translating...", style = MaterialTheme.typography.bodyMedium)
                                    }
                                }
                            }
                        }
                    }

                    Spacer(modifier = Modifier.height(24.dp))

                    // 4. Two-Way Microphone Controls
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceEvenly,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        // TALK BUTTON (Associate -> Customer)
                        val isTalkActive = isRecording && activeMode == RecordingMode.TALK
                        MicrophoneButton(
                            label = "TALK",
                            subLabel = "Translate to Customer",
                            isActive = isTalkActive,
                            containerColor = GTasksTheme.colors.colorPrimary,
                            onClick = {
                                if (isTalkActive) {
                                    viewModel.stopRecordingAndTranslate(targetLanguage)
                                } else if (!isRecording) {
                                    viewModel.startRecording(RecordingMode.TALK)
                                }
                            }
                        )

                        // LISTEN BUTTON (Customer -> Associate)
                        val isListenActive = isRecording && activeMode == RecordingMode.LISTEN
                        MicrophoneButton(
                            label = "LISTEN",
                            subLabel = "Translate to English",
                            isActive = isListenActive,
                            containerColor = GTasksTheme.colors.colorSecondary,
                            onClick = {
                                if (isListenActive) {
                                    viewModel.stopRecordingAndTranslate(targetLanguage, customerGender)
                                } else if (!isRecording) {
                                    viewModel.startRecording(RecordingMode.LISTEN)
                                }
                            }
                        )
                    }
                }
            } else {
                // TAB 2: VOICE & PROFILE SETTINGS
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .padding(16.dp)
                ) {
                    Text(
                        text = "Voice Customization Settings",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = GTasksTheme.colors.textPrimary,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    // Profile Details Card
                    Card(
                        colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                        border = BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 16.dp)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(text = "Associate Profile", style = MaterialTheme.typography.labelSmall, color = GTasksTheme.colors.textMuted)
                            Text(text = profile?.name ?: "Unknown", style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold))
                            Text(text = profile?.email ?: "", style = MaterialTheme.typography.bodySmall, color = GTasksTheme.colors.textSecondary)
                        }
                    }

                    // Voice Gender Segment Selector
                    Card(
                        colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                        border = BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 16.dp)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(
                                text = "Your Voice Gender Preference",
                                style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                color = GTasksTheme.colors.textPrimary
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Row(modifier = Modifier.fillMaxWidth()) {
                                FilterChip(
                                    selected = profile?.voiceGenderPreference == "FEMALE",
                                    onClick = { viewModel.updateVoicePreferences("FEMALE", profile?.voiceNamePreference ?: "en-US-Journey-F") },
                                    label = { Text("Female Style") },
                                    modifier = Modifier.weight(1f).padding(end = 8.dp)
                                )
                                FilterChip(
                                    selected = profile?.voiceGenderPreference == "MALE",
                                    onClick = { viewModel.updateVoicePreferences("MALE", profile?.voiceNamePreference ?: "en-US-Journey-D") },
                                    label = { Text("Male Style") },
                                    modifier = Modifier.weight(1f)
                                )
                            }
                        }
                    }

                    // Premium Voice Selection Dropdown Card
                    Card(
                        colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                        border = BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 24.dp)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(
                                text = "Select Voice Style (HD Premium)",
                                style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                color = GTasksTheme.colors.textPrimary
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Box(modifier = Modifier.fillMaxWidth()) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .border(1.dp, GTasksTheme.colors.borderMuted, RoundedCornerShape(8.dp))
                                        .clickable { voiceDropdownExpanded = true }
                                        .padding(12.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.SpaceBetween
                                ) {
                                    Text(
                                        text = (profile?.voiceNamePreference ?: "en-US-Journey-F").substringAfterLast('-'),
                                        style = MaterialTheme.typography.bodyMedium,
                                        color = GTasksTheme.colors.textPrimary
                                    )
                                    Icon(imageVector = Icons.Default.ArrowDropDown, contentDescription = null)
                                }

                                DropdownMenu(
                                    expanded = voiceDropdownExpanded,
                                    onDismissRequest = { voiceDropdownExpanded = false },
                                    modifier = Modifier
                                        .fillMaxWidth(0.9f)
                                        .background(GTasksTheme.colors.bgMain)
                                        .border(1.dp, GTasksTheme.colors.borderMuted)
                                ) {
                                    // List only premium voices matching selected gender preference
                                    val filteredVoices = voices.filter { it.gender == (profile?.voiceGenderPreference ?: "FEMALE") }
                                    if (filteredVoices.isEmpty()) {
                                        DropdownMenuItem(
                                            text = { Text("No voices loaded") },
                                            onClick = {}
                                        )
                                    } else {
                                        filteredVoices.forEach { v ->
                                            DropdownMenuItem(
                                                text = {
                                                    Row(
                                                        modifier = Modifier.fillMaxWidth(),
                                                        horizontalArrangement = Arrangement.SpaceBetween,
                                                        verticalAlignment = Alignment.CenterVertically
                                                    ) {
                                                        Row(
                                                            verticalAlignment = Alignment.CenterVertically,
                                                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                                                        ) {
                                                            // Play preview button
                                                            IconButton(
                                                                onClick = {
                                                                    viewModel.playVoicePreview(v.name, v.languageCode)
                                                                },
                                                                modifier = Modifier.size(28.dp)
                                                            ) {
                                                                Icon(
                                                                    imageVector = Icons.Default.PlayArrow,
                                                                    contentDescription = "Test Voice",
                                                                    tint = GTasksTheme.colors.colorPrimary,
                                                                    modifier = Modifier.size(18.dp)
                                                                )
                                                            }
                                                            Text(
                                                                text = v.name.substringAfterLast('-'),
                                                                style = MaterialTheme.typography.bodyMedium
                                                            )
                                                        }
                                                        Text(
                                                            text = v.qualityClass,
                                                            style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                                                            color = GTasksTheme.colors.colorAccent
                                                        )
                                                    }
                                                },
                                                onClick = {
                                                    viewModel.updateVoicePreferences(profile?.voiceGenderPreference ?: "FEMALE", v.name)
                                                    voiceDropdownExpanded = false
                                                }
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }

                    // 3. Premium Custom Voice Cloning Section!
                    Card(
                        colors = CardDefaults.cardColors(containerColor = GTasksTheme.colors.bgCard),
                        border = BorderStroke(1.dp, GTasksTheme.colors.borderMuted),
                        modifier = Modifier
                            .fillMaxWidth()
                            .weight(1f)
                    ) {
                        Column(
                            modifier = Modifier
                                .fillMaxSize()
                                .padding(16.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            Text(
                                text = "Create Your Custom Voice Clone",
                                style = MaterialTheme.typography.bodyLarge.copy(fontWeight = FontWeight.Black),
                                color = GTasksTheme.colors.textPrimary
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = "Speak the required consent phrase to train Hanna to translate using your own custom cloned voice:",
                                style = MaterialTheme.typography.bodySmall,
                                color = GTasksTheme.colors.textSecondary,
                                textAlign = TextAlign.Center,
                                modifier = Modifier.padding(horizontal = 8.dp)
                            )
                            Spacer(modifier = Modifier.height(12.dp))
                            
                            // Authorization phrase card
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(GTasksTheme.colors.borderMuted.copy(alpha = 0.2f), RoundedCornerShape(8.dp))
                                    .border(1.dp, GTasksTheme.colors.borderMuted, RoundedCornerShape(8.dp))
                                    .padding(12.dp)
                            ) {
                                val consentText = remember(profile?.voiceNamePreference) {
                                    val voicePref = profile?.voiceNamePreference?.lowercase() ?: "en-us"
                                    when {
                                        voicePref.startsWith("es") -> "Soy el propietario de esta voz y doy mi consentimiento para que Google la utilice para crear un modelo de voz sintética."
                                        voicePref.startsWith("fr") -> "Je suis le propriétaire de cette voix et j'autorise Google à utiliser cette voix pour créer un modèle de voix synthétique."
                                        voicePref.startsWith("de") -> "Ich bin der Eigentümer dieser Stimme und bin damit einverstanden, dass Google diese Stimme zur Erstellung eines synthetischen Stimmmodells verwendet."
                                        voicePref.startsWith("it") -> "Sono il proprietario di questa voce e acconsento che Google la utilizzi per creare un modelo di voce sintetica."
                                        voicePref.startsWith("ja") -> "私はこの音声の所有者であり、Googleがこの音声を使用して音声合成モデルを作成することを承認します。"
                                        voicePref.startsWith("ko") -> "나는 이 음성의 소유자이며 구글이 이 음성을 사용하여 음성 합성 모델을 생성할 것을 허용합니다。"
                                        else -> "I am the owner of this voice and I consent to Google using this voice to create a synthetic voice model."
                                    }
                                }

                                Text(
                                    text = "\"$consentText\"",
                                    style = MaterialTheme.typography.bodyMedium.copy(
                                        fontStyle = FontStyle.Italic,
                                        fontWeight = FontWeight.Bold,
                                        letterSpacing = 0.5.sp
                                    ),
                                    color = GTasksTheme.colors.colorPrimary,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.fillMaxWidth()
                                )
                            }
                            
                            Spacer(modifier = Modifier.height(16.dp))

                            val isCloningActive = isRecording && activeMode == RecordingMode.TALK
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(16.dp)
                            ) {
                                // Record button
                                Button(
                                    onClick = {
                                        if (isCloningActive) {
                                            viewModel.stopRecordingOnly()
                                        } else {
                                            viewModel.startRecording(RecordingMode.TALK)
                                        }
                                    },
                                    colors = ButtonDefaults.buttonColors(
                                        containerColor = if (isCloningActive) GTasksTheme.colors.colorCritical else GTasksTheme.colors.colorPrimary
                                    )
                                ) {
                                    Icon(
                                        imageVector = if (isCloningActive) Icons.Default.Stop else Icons.Default.Mic,
                                        contentDescription = null
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(text = if (isCloningActive) "Stop Recording" else "Record Consent")
                                }

                                // Train button
                                Button(
                                    onClick = {
                                        viewModel.cloneVoice {
                                            // Handle success callback!
                                        }
                                    },
                                    enabled = !isRecording,
                                    colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorSecondary)
                                ) {
                                    Icon(imageVector = Icons.Default.AutoAwesome, contentDescription = null)
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(text = "Generate Voice Clone")
                                }
                            }

                            Spacer(modifier = Modifier.height(16.dp))

                            // Cloned voice status indicator
                            if (profile?.clonedVoiceKey?.isNotEmpty() == true) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    modifier = Modifier.padding(top = 8.dp)
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Check,
                                        tint = Color.Green,
                                        modifier = Modifier.size(18.dp),
                                        contentDescription = null
                                    )
                                    Spacer(modifier = Modifier.width(6.dp))
                                    Text(
                                        text = "Voice Clone Activated!",
                                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                                        color = Color.Green
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

// Sub-components

@Composable
fun MicrophoneButton(
    label: String,
    subLabel: String,
    isActive: Boolean,
    containerColor: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    // Pulse Animation for active recording states!
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val scale by infiniteTransition.animateFloat(
        initialValue = 1.0f,
        targetValue = 1.2f,
        animationSpec = infiniteRepeatable(
            animation = tween(800, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "scale"
    )

    val colorAccent = if (isActive) GTasksTheme.colors.colorCritical else containerColor

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = modifier.width(160.dp)
    ) {
        Box(
            contentAlignment = Alignment.Center,
            modifier = Modifier
                .scale(if (isActive) scale else 1.0f)
                .size(72.dp)
                .background(
                    color = colorAccent.copy(alpha = if (isActive) 0.15f else 0.1f),
                    shape = CircleShape
                )
                .border(
                    width = 2.dp,
                    color = colorAccent.copy(alpha = if (isActive) 0.8f else 0.3f),
                    shape = CircleShape
                )
                .clip(CircleShape)
                .clickable { onClick() }
        ) {
            Icon(
                imageVector = if (isActive) Icons.Default.Stop else Icons.Default.Mic,
                contentDescription = label,
                tint = colorAccent,
                modifier = Modifier.size(36.dp)
            )
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Black),
            color = GTasksTheme.colors.textPrimary
        )
        Text(
            text = subLabel,
            style = MaterialTheme.typography.labelSmall,
            color = GTasksTheme.colors.textMuted,
            textAlign = TextAlign.Center
        )
    }
}
