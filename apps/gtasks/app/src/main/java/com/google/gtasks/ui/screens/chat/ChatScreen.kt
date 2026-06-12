package com.google.gtasks.ui.screens.chat

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.google.gtasks.ui.a2ui.ButtonAction
import kotlinx.serialization.json.JsonElement
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.google.gtasks.ui.a2ui.A2UIRenderer
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatScreen(
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: ChatViewModel = viewModel(factory = ChatViewModel.Factory)
) {
    val messages by viewModel.messages.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    
    val useLocalReasoning by viewModel.useLocalReasoning.collectAsState()
    val isLocalReady by viewModel.isLocalGemmaReady.collectAsState()
    val localStatus by viewModel.localGemmaStatus.collectAsState()

    var textInput by remember { mutableStateOf("") }
    val listState = rememberLazyListState()

    // Auto-scroll to latest message on new items
    LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) {
            listState.animateScrollToItem(messages.size - 1)
        }
    }

    Scaffold(
        topBar = {
            // Top App Bar
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .glassmorphic(shape = RoundedCornerShape(0.dp), elevation = 4.dp)
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .statusBarsPadding()
                        .padding(horizontal = 12.dp, vertical = 10.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        IconButton(onClick = onBackClick) {
                            Icon(
                                imageVector = Icons.Default.ArrowBack,
                                contentDescription = "Back to list",
                                tint = GTasksTheme.colors.textPrimary
                            )
                        }
                        
                        Column(modifier = Modifier.padding(start = 4.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text(
                                    text = "HANNA COACH",
                                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                                    color = GTasksTheme.colors.textPrimary
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                // Glowing Status Dot
                                val dotColor = when {
                                    !useLocalReasoning -> GTasksTheme.colors.colorPrimary // Cloud (Indigo)
                                    isLocalReady -> GTasksTheme.colors.colorAccent     // Local Ready (Teal)
                                    else -> GTasksTheme.colors.colorCritical          // Local Not Ready (Red)
                                }
                                Box(
                                    modifier = Modifier
                                        .size(8.dp)
                                        .background(dotColor, CircleShape)
                                )
                            }
                            Text(
                                text = if (useLocalReasoning) "Local Gemma 2B" else "Cloud Gemini Grounded",
                                style = MaterialTheme.typography.labelSmall,
                                color = GTasksTheme.colors.textSecondary
                            )
                        }
                    }

                    // Toggle Reasoning Button
                    IconButton(
                        onClick = { viewModel.toggleReasoningMode() },
                        colors = IconButtonDefaults.iconButtonColors(contentColor = GTasksTheme.colors.textSecondary)
                    ) {
                        Icon(
                            imageVector = Icons.Default.AutoAwesome,
                            contentDescription = "Toggle reasoning engine",
                            tint = if (useLocalReasoning) GTasksTheme.colors.colorAccent else GTasksTheme.colors.textSecondary
                        )
                    }
                }
            }
        },
        bottomBar = {
            // Chat Input Bar
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .imePadding()
                    .navigationBarsPadding()
                    .glassmorphic(shape = RoundedCornerShape(0.dp), elevation = 6.dp)
            ) {
                Column {
                    // Thinking indicator
                    AnimatedVisibility(visible = isLoading) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 20.dp, vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(14.dp),
                                strokeWidth = 2.dp,
                                color = GTasksTheme.colors.colorSecondary
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = if (useLocalReasoning) "Gemma is reasoning locally..." else "Hanna is drafting grounded A2UI cards...",
                                style = MaterialTheme.typography.labelSmall,
                                color = GTasksTheme.colors.textSecondary
                            )
                        }
                    }

                    // Text Field Row
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        OutlinedTextField(
                            value = textInput,
                            onValueChange = { textInput = it },
                            placeholder = { Text("Ask Hanna or trigger local checks...", color = GTasksTheme.colors.textMuted) },
                            modifier = Modifier.weight(1f),
                            maxLines = 3,
                            shape = RoundedCornerShape(24.dp),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = GTasksTheme.colors.colorPrimary,
                                unfocusedBorderColor = GTasksTheme.colors.borderMuted,
                                focusedTextColor = GTasksTheme.colors.textPrimary,
                                unfocusedTextColor = GTasksTheme.colors.textPrimary
                            )
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        IconButton(
                            onClick = {
                                if (textInput.isNotBlank()) {
                                    viewModel.sendMessage(textInput)
                                    textInput = ""
                                }
                            },
                            enabled = textInput.isNotBlank() && !isLoading,
                            colors = IconButtonDefaults.iconButtonColors(
                                containerColor = GTasksTheme.colors.colorPrimary,
                                contentColor = Color.White,
                                disabledContainerColor = GTasksTheme.colors.textMuted.copy(alpha = 0.2f)
                            ),
                            modifier = Modifier.size(48.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.Send,
                                contentDescription = "Send prompt"
                            )
                        }
                    }
                }
            }
        },
        containerColor = GTasksTheme.colors.bgMain,
        modifier = modifier
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
                .background(
                    Brush.radialGradient(
                        colors = listOf(
                            Color(0x1F8B5CF6), // Translucent Violet
                            GTasksTheme.colors.bgMain
                        ),
                        radius = 1000f
                    )
                )
        ) {
            Column(modifier = Modifier.fillMaxSize()) {
                // Show model missing guidance bar if local reasoning enabled but model not found
                if (useLocalReasoning && !isLocalReady) {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(GTasksTheme.colors.colorCritical.copy(alpha = 0.1f))
                            .border(1.dp, GTasksTheme.colors.colorCritical.copy(alpha = 0.2f))
                            .padding(12.dp)
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(
                                imageVector = Icons.Default.Info,
                                contentDescription = "Gemma Offline Info",
                                tint = GTasksTheme.colors.colorCritical,
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = localStatus,
                                style = MaterialTheme.typography.bodySmall.copy(fontSize = 11.sp),
                                color = GTasksTheme.colors.textPrimary
                            )
                        }
                    }
                }

                // Chat Stream
                LazyColumn(
                    state = listState,
                    contentPadding = PaddingValues(horizontal = 20.dp, vertical = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                    modifier = Modifier.weight(1f)
                ) {
                    items(messages) { message ->
                        ChatBubble(
                            message = message,
                            onA2UIAction = { action, data ->
                                viewModel.handleA2UIAction(action, data)
                            }
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun ChatBubble(
    message: ChatMessage,
    onA2UIAction: (ButtonAction, Map<String, JsonElement>) -> Unit,
    modifier: Modifier = Modifier
) {
    when (message.role) {
        "user" -> {
            // User bubble: Right aligned
            Box(
                modifier = modifier
                    .fillMaxWidth(),
                contentAlignment = Alignment.CenterEnd
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(0.8f)
                        .background(GTasksTheme.colors.bgCardHover, RoundedCornerShape(16.dp, 16.dp, 0.dp, 16.dp))
                        .border(1.dp, GTasksTheme.colors.borderMuted, RoundedCornerShape(16.dp, 16.dp, 0.dp, 16.dp))
                        .padding(14.dp)
                ) {
                    Text(
                        text = message.content,
                        color = GTasksTheme.colors.textPrimary,
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }
        }
        "assistant" -> {
            // Hanna Coach bubble: Left aligned
            Column(
                modifier = modifier.fillMaxWidth(),
                horizontalAlignment = Alignment.Start
            ) {
                // Coach Text
                Box(
                    modifier = Modifier
                        .fillMaxWidth(0.85f)
                        .background(GTasksTheme.colors.colorSecondary.copy(alpha = 0.15f), RoundedCornerShape(16.dp, 16.dp, 16.dp, 0.dp))
                        .border(1.dp, GTasksTheme.colors.colorSecondary.copy(alpha = 0.3f), RoundedCornerShape(16.dp, 16.dp, 16.dp, 0.dp))
                        .padding(14.dp)
                ) {
                    Text(
                        text = message.content,
                        color = GTasksTheme.colors.textPrimary,
                        style = MaterialTheme.typography.bodyMedium
                    )
                }

                // Native A2UI Card Embedded Directly In The Message Stream!
                message.transaction?.let { transaction ->
                    Spacer(modifier = Modifier.height(10.dp))
                    A2UIRenderer(
                        transaction = transaction,
                        modifier = Modifier
                            .fillMaxWidth(0.9f)
                            .padding(start = 6.dp),
                        onAction = onA2UIAction
                    )
                }
            }
        }
        "system" -> {
            // System Audit message: Centered
            Box(
                modifier = modifier
                    .fillMaxWidth()
                    .padding(vertical = 4.dp),
                contentAlignment = Alignment.Center
            ) {
                Box(
                    modifier = Modifier
                        .background(GTasksTheme.colors.colorPrimary.copy(alpha = 0.05f), RoundedCornerShape(8.dp))
                        .border(1.dp, GTasksTheme.colors.borderGlow, RoundedCornerShape(8.dp))
                        .padding(horizontal = 16.dp, vertical = 8.dp)
                ) {
                    Text(
                        text = message.content,
                        color = GTasksTheme.colors.colorPrimary,
                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                        textAlign = androidx.compose.ui.text.style.TextAlign.Center
                    )
                }
            }
        }
    }
}
