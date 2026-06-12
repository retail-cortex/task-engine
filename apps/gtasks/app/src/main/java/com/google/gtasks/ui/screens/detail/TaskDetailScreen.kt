package com.google.gtasks.ui.screens.detail

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.LockOpen
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.Place
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.filled.SwapHoriz
import androidx.lifecycle.viewmodel.compose.viewModel
import com.google.gtasks.data.model.ChecklistItem
import com.google.gtasks.data.model.TaskExecution
import com.google.gtasks.ui.screens.tasks.Colleague
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskDetailScreen(
    taskId: String,
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: TaskDetailViewModel = viewModel(factory = TaskDetailViewModel.Factory)
) {
    val uiState by viewModel.uiState.collectAsState()
    val isCompletable by viewModel.isCompletable.collectAsState()
    val isSubmitting by viewModel.isSubmitting.collectAsState()

    var justificationText by remember { mutableStateOf("") }

    // Initial load
    LaunchedEffect(taskId) {
        viewModel.loadTaskDetails(taskId)
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
                        .padding(horizontal = 12.dp, vertical = 14.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    IconButton(onClick = onBackClick) {
                        Icon(
                            imageVector = Icons.Default.ArrowBack,
                            contentDescription = "Go Back",
                            tint = GTasksTheme.colors.textPrimary
                        )
                    }
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "TASK COMPLIANCE DETAILS",
                        style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                        color = GTasksTheme.colors.textPrimary
                    )
                }
            }
        },
        bottomBar = {
            // Bottom-pinned Complete Task Button
            val state = uiState
            if (state is TaskDetailUiState.Success) {
                val task = state.task
                if (task.status != "COMPLETED") {
                    val isCompletable = viewModel.isCompletable.value
                    val isSubmitting by viewModel.isSubmitting.collectAsState()
                    val currentUserId = viewModel.currentUserId
                    var showTradeDialog by remember { mutableStateOf(false) }
                    
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .navigationBarsPadding()
                            .padding(horizontal = 20.dp, vertical = 16.dp)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            // Offer Trade Button (Only if assigned to current user!)
                            if (task.assigneeID == currentUserId) {
                                OutlinedButton(
                                    onClick = { showTradeDialog = true },
                                    enabled = !isSubmitting,
                                    colors = ButtonDefaults.outlinedButtonColors(contentColor = GTasksTheme.colors.colorPrimary),
                                    border = BorderStroke(1.dp, GTasksTheme.colors.colorPrimary.copy(alpha = 0.6f)),
                                    shape = RoundedCornerShape(12.dp),
                                    modifier = Modifier
                                        .weight(0.4f)
                                        .height(56.dp)
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.SwapHoriz,
                                        contentDescription = "Trade Task"
                                    )
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text("Trade", fontWeight = FontWeight.Bold)
                                }
                            }
                            
                            // Complete Task Button
                            Button(
                                onClick = {
                                    viewModel.completeTask(taskId, onBackClick)
                                },
                                enabled = isCompletable && !isSubmitting,
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = GTasksTheme.colors.colorAccent,
                                    disabledContainerColor = GTasksTheme.colors.textMuted.copy(alpha = 0.2f)
                                ),
                                shape = RoundedCornerShape(12.dp),
                                modifier = Modifier
                                    .weight(if (task.assigneeID == currentUserId) 0.6f else 1f)
                                    .height(56.dp)
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.Center
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Check,
                                        contentDescription = "Success check",
                                        tint = if (isCompletable) Color.White else GTasksTheme.colors.textMuted
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = if (isSubmitting) "SYNCHRONIZING..." else "COMPLETE RUN",
                                        color = if (isCompletable) Color.White else GTasksTheme.colors.textMuted,
                                        fontWeight = FontWeight.Bold,
                                        style = MaterialTheme.typography.labelLarge,
                                        maxLines = 1,
                                        overflow = TextOverflow.Ellipsis
                                    )
                                }
                            }
                        }
                        
                        if (showTradeDialog) {
                            TradeProposeDialog(
                                taskName = task.task?.name ?: "Task",
                                colleagues = viewModel.colleagues,
                                onDismiss = { showTradeDialog = false },
                                onSelectColleague = { colleagueId ->
                                    viewModel.proposeTrade(task.id, colleagueId) {
                                        showTradeDialog = false
                                    }
                                }
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
                            Color(0x1FEF4444), // Translucent Red (Critical/Warning Glow)
                            GTasksTheme.colors.bgMain
                        ),
                        radius = 1000f
                    )
                )
        ) {
            when (val state = uiState) {
                is TaskDetailUiState.Loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = GTasksTheme.colors.colorPrimary)
                    }
                }
                is TaskDetailUiState.Success -> {
                    val task = state.task
                    
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(horizontal = 20.dp)
                            .verticalScroll(rememberScrollState())
                    ) {
                        Spacer(modifier = Modifier.height(20.dp))

                        // 1. Task Summary Header Card
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .glassmorphic(shape = RoundedCornerShape(20.dp))
                        ) {
                            Column(modifier = Modifier.padding(20.dp)) {
                                val priorityLabel = when (task.priority) {
                                    1 -> "HIGH PRIORITY"
                                    2 -> "MEDIUM PRIORITY"
                                    else -> "LOW PRIORITY"
                                }
                                val priorityColor = when (task.priority) {
                                    1 -> GTasksTheme.colors.colorCritical
                                    2 -> GTasksTheme.colors.colorPrimary
                                    else -> GTasksTheme.colors.colorAccent
                                }
                                
                                Text(
                                    text = priorityLabel,
                                    color = priorityColor,
                                    style = MaterialTheme.typography.labelMedium.copy(fontWeight = FontWeight.Bold)
                                )

                                Text(
                                    text = task.task.name,
                                    style = MaterialTheme.typography.displayMedium.copy(fontSize = 24.sp, fontWeight = FontWeight.Black),
                                    color = GTasksTheme.colors.textPrimary,
                                    modifier = Modifier.padding(top = 8.dp)
                                )

                                Text(
                                    text = task.description ?: task.task.description ?: "No description provided.",
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = GTasksTheme.colors.textSecondary,
                                    modifier = Modifier.padding(top = 8.dp)
                                )

                                Spacer(modifier = Modifier.height(14.dp))
                                
                                // Interactive Location Map Trigger Row
                                var showMapDialog by remember { mutableStateOf(false) }
                                val resolvedLocation = resolveLocationName(task.task.name)
                                
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    modifier = Modifier
                                        .background(GTasksTheme.colors.bgInput, RoundedCornerShape(8.dp))
                                        .border(BorderStroke(1.dp, GTasksTheme.colors.borderMuted), RoundedCornerShape(8.dp))
                                        .clickable { showMapDialog = true }
                                        .padding(horizontal = 12.dp, vertical = 8.dp)
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Place,
                                        contentDescription = "Map Location Icon",
                                        tint = GTasksTheme.colors.colorAccent,
                                        modifier = Modifier.size(16.dp)
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = "Location: $resolvedLocation",
                                        style = MaterialTheme.typography.bodySmall.copy(fontWeight = FontWeight.Bold),
                                        color = GTasksTheme.colors.textPrimary
                                    )
                                    Spacer(modifier = Modifier.width(12.dp))
                                    Text(
                                        text = "(View Floorplan)",
                                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                                        color = GTasksTheme.colors.colorPrimary
                                    )
                                }

                                if (showMapDialog) {
                                    StoreMapDialog(
                                        siteId = viewModel.siteId,
                                        siteName = viewModel.siteName,
                                        taskName = task.task.name,
                                        onDismiss = { showMapDialog = false }
                                    )
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(24.dp))

                        // 2. Constraints / Asset Bypass Block
                        // Render a warning and override form if there's a hard blocker constraint
                        // In this task engine, Cash Drops and Chiller checks are typical override targets.
                        val isCashDrop = task.task.name.contains("Cash Drop", ignoreCase = true) || task.task.id.contains("d000fa44")
                        if (isCashDrop && task.status != "COMPLETED") {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .glassmorphic(shape = RoundedCornerShape(16.dp))
                                    .border(BorderStroke(1.dp, GTasksTheme.colors.colorWarning.copy(alpha = 0.3f)), RoundedCornerShape(16.dp))
                            ) {
                                Column(modifier = Modifier.padding(16.dp)) {
                                    Row(verticalAlignment = Alignment.CenterVertically) {
                                        Icon(
                                            imageVector = Icons.Default.Warning,
                                            contentDescription = "Warning",
                                            tint = GTasksTheme.colors.colorWarning
                                        )
                                        Spacer(modifier = Modifier.width(8.dp))
                                        Text(
                                            text = "SECURITY CONSTRAINT LOCK",
                                            style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                                            color = GTasksTheme.colors.colorWarning
                                        )
                                    }
                                    Text(
                                        text = "Register cash drawer limits exceeded. Requires vault drop verification. Submit supervisor override justification to bypass.",
                                        style = MaterialTheme.typography.bodySmall,
                                        color = GTasksTheme.colors.textSecondary,
                                        modifier = Modifier.padding(top = 8.dp, bottom = 16.dp)
                                    )

                                    OutlinedTextField(
                                        value = justificationText,
                                        onValueChange = { justificationText = it },
                                        label = { Text("Supervisor Justification Code") },
                                        modifier = Modifier.fillMaxWidth(),
                                        colors = OutlinedTextFieldDefaults.colors(
                                            focusedBorderColor = GTasksTheme.colors.colorWarning,
                                            unfocusedBorderColor = GTasksTheme.colors.borderMuted
                                        )
                                    )

                                    Spacer(modifier = Modifier.height(12.dp))

                                    Button(
                                        onClick = {
                                            viewModel.submitAssetOverride(
                                                executionId = taskId,
                                                assetId = "CASH-CEILING-DRAWER-4",
                                                justification = justificationText
                                            )
                                            justificationText = ""
                                        },
                                        enabled = justificationText.isNotEmpty() && !isSubmitting,
                                        colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorWarning),
                                        modifier = Modifier.align(Alignment.End)
                                    ) {
                                        Icon(Icons.Default.LockOpen, "Lock Icon", modifier = Modifier.size(16.dp))
                                        Spacer(modifier = Modifier.width(8.dp))
                                        Text("Authorize Override")
                                    }
                                }
                            }
                            Spacer(modifier = Modifier.height(24.dp))
                        }

                        // 3. Interactive Checklist Title
                        Text(
                            text = "COMPLIANCE CHECKLIST STEPS",
                            style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold),
                            color = GTasksTheme.colors.textSecondary,
                            modifier = Modifier.padding(bottom = 12.dp)
                        )

                        // Checklist Steps
                        if (viewModel.checklist.isEmpty()) {
                            Text(
                                text = "No checklist steps specified.",
                                color = GTasksTheme.colors.textMuted,
                                style = MaterialTheme.typography.bodyMedium,
                                modifier = Modifier.padding(vertical = 8.dp)
                            )
                        } else {
                            viewModel.checklist.forEach { item ->
                                ChecklistStepItem(
                                    item = item,
                                    onClick = {
                                        if (task.status != "COMPLETED") {
                                            viewModel.toggleChecklistItem(taskId, item.step)
                                        }
                                    },
                                    enabled = task.status != "COMPLETED"
                                )
                                Spacer(modifier = Modifier.height(10.dp))
                            }
                        }

                        Spacer(modifier = Modifier.height(100.dp)) // bottom padding
                    }
                }
                is TaskDetailUiState.Error -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Text(
                                text = state.message,
                                color = GTasksTheme.colors.colorCritical,
                                style = MaterialTheme.typography.bodyLarge,
                                textAlign = TextAlign.Center
                            )
                            Spacer(modifier = Modifier.height(16.dp))
                            Button(
                                onClick = { viewModel.loadTaskDetails(taskId) },
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
}

@Composable
private fun ChecklistStepItem(
    item: ChecklistItem,
    onClick: () -> Unit,
    enabled: Boolean,
    modifier: Modifier = Modifier
) {
    val themeColors = GTasksTheme.colors
    val itemBorderColor = if (item.completed) themeColors.colorPrimary.copy(alpha = 0.4f) else themeColors.borderMuted

    Box(
        modifier = modifier
            .fillMaxWidth()
            .glassmorphic(shape = RoundedCornerShape(12.dp), elevation = 2.dp)
            .border(BorderStroke(1.dp, itemBorderColor), RoundedCornerShape(12.dp))
            .clickable(enabled = enabled, onClick = onClick)
            .padding(14.dp)
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.fillMaxWidth()
        ) {
            Checkbox(
                checked = item.completed,
                onCheckedChange = { onClick() },
                enabled = enabled,
                colors = CheckboxDefaults.colors(
                    checkedColor = themeColors.colorPrimary,
                    uncheckedColor = themeColors.textMuted
                )
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = item.action,
                    color = if (item.completed) themeColors.textSecondary else themeColors.textPrimary,
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontWeight = if (item.required) FontWeight.Bold else FontWeight.Normal
                    )
                )
                if (item.required && !item.completed) {
                    Text(
                        text = "Required Step",
                        color = themeColors.colorCritical.copy(alpha = 0.8f),
                        style = MaterialTheme.typography.labelSmall,
                        modifier = Modifier.padding(top = 2.dp)
                    )
                }
            }
        }
    }
}

@Composable
private fun TradeProposeDialog(
    taskName: String,
    colleagues: List<Colleague>,
    onDismiss: () -> Unit,
    onSelectColleague: (String) -> Unit
) {
    Dialog(onDismissRequest = onDismiss) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .glassmorphic(shape = RoundedCornerShape(24.dp), elevation = 8.dp)
                .padding(24.dp)
        ) {
            Column {
                Text(
                    text = "Propose Task Trade",
                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                    color = GTasksTheme.colors.textPrimary
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "Select a colleague to offer coverage for:\n$taskName",
                    style = MaterialTheme.typography.bodySmall,
                    color = GTasksTheme.colors.textSecondary
                )
                Spacer(modifier = Modifier.height(16.dp))

                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.heightIn(max = 240.dp)
                ) {
                    items(colleagues) { colleague ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(GTasksTheme.colors.bgInput, RoundedCornerShape(8.dp))
                                .border(BorderStroke(1.dp, GTasksTheme.colors.borderMuted), RoundedCornerShape(8.dp))
                                .clickable { onSelectColleague(colleague.id) }
                                .padding(horizontal = 14.dp, vertical = 10.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Column {
                                Text(colleague.name, style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold), color = GTasksTheme.colors.textPrimary)
                                Text(colleague.roleName, style = MaterialTheme.typography.labelSmall, color = GTasksTheme.colors.textMuted)
                            }
                            Icon(androidx.compose.material.icons.Icons.Default.ChevronRight, contentDescription = null, tint = GTasksTheme.colors.textMuted)
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))
                TextButton(onClick = onDismiss, modifier = Modifier.align(Alignment.End)) {
                    Text("Cancel", color = GTasksTheme.colors.textSecondary)
                }
            }
        }
    }
}
