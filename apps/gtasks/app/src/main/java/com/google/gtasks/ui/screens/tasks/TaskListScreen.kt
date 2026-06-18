package com.google.gtasks.ui.screens.tasks

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.ExitToApp
import androidx.compose.material.icons.filled.FilterList
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material.icons.filled.Sort
import androidx.compose.material.icons.filled.SwapHoriz
import androidx.compose.material.icons.filled.Translate
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Cancel
import androidx.compose.material.icons.filled.Storefront
import androidx.compose.material3.*
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
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
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.viewmodel.compose.viewModel
import com.google.gtasks.data.model.TaskExecution
import com.google.gtasks.data.model.Trade
import com.google.gtasks.ui.theme.GTasksTheme
import com.google.gtasks.ui.theme.glassmorphic

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskListScreen(
    onTaskClick: (String) -> Unit,
    onChatClick: () -> Unit,
    onTranslateClick: () -> Unit,
    onLogout: () -> Unit,
    onChangeSite: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: TaskListViewModel = viewModel(factory = TaskListViewModel.Factory)
) {
    val uiState by viewModel.uiState.collectAsState()
    val colleagues by viewModel.colleagues.collectAsState()

    // Calculate active item counts for badges
    val (assignedCount, availableCount, pendingTradesCount) = when (val state = uiState) {
        is TaskListUiState.Success -> Triple(
            state.assignedTasks.size,
            state.availableTasks.size,
            state.pendingTrades.size
        )
        else -> Triple(0, 0, 0)
    }

    val statusFilter by viewModel.statusFilter.collectAsState()
    val priorityFilter by viewModel.priorityFilter.collectAsState()
    val sortBy by viewModel.sortBy.collectAsState()

    var selectedTabIndex by remember { mutableIntStateOf(0) }
    var activeTradeTask by remember { mutableStateOf<TaskExecution?>(null) }

    // Dynamically refresh data on ON_RESUME (e.g. popBackStack or app foreground)
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) {
                viewModel.loadData()
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }

    Scaffold(
        topBar = {
            // 1. Premium Glassmorphic Header Banner showing active Site and derived Role
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .glassmorphic(shape = RoundedCornerShape(0.dp), elevation = 4.dp)
            ) {
                Column {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .statusBarsPadding()
                            .padding(horizontal = 20.dp, vertical = 16.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Column {
                            Text(
                                text = "GEMINI TASKS",
                                style = MaterialTheme.typography.titleLarge.copy(
                                    fontWeight = FontWeight.Black,
                                    letterSpacing = 1.sp
                                ),
                                color = GTasksTheme.colors.textPrimary
                            )
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                modifier = Modifier.padding(top = 4.dp)
                            ) {
                                Text(
                                    text = viewModel.activeSiteName,
                                    style = MaterialTheme.typography.labelMedium,
                                    color = GTasksTheme.colors.colorPrimary,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(modifier = Modifier.width(6.dp))
                                Box(
                                    modifier = Modifier
                                        .size(4.dp)
                                        .background(GTasksTheme.colors.textMuted, RoundedCornerShape(50))
                                )
                                Spacer(modifier = Modifier.width(6.dp))
                                // Derived Role display!
                                Text(
                                    text = viewModel.activeRoleName,
                                    style = MaterialTheme.typography.labelSmall,
                                    color = GTasksTheme.colors.colorAccent,
                                    fontWeight = FontWeight.Bold
                                )
                            }
                        }

                        Row(
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            IconButton(
                                onClick = onChangeSite,
                                colors = IconButtonDefaults.iconButtonColors(contentColor = GTasksTheme.colors.textSecondary)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Storefront,
                                    contentDescription = "Switch Site"
                                )
                            }
                            IconButton(
                                onClick = onLogout,
                                colors = IconButtonDefaults.iconButtonColors(contentColor = GTasksTheme.colors.textSecondary)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.ExitToApp,
                                    contentDescription = "Logout"
                                )
                            }
                        }
                    }

                    // 2. Navigation Tabs (My Work, Available, Trades)
                    TabRow(
                        selectedTabIndex = selectedTabIndex,
                        containerColor = Color.Transparent,
                        contentColor = GTasksTheme.colors.textPrimary,
                        indicator = { tabPositions ->
                            TabRowDefaults.SecondaryIndicator(
                                Modifier.tabIndicatorOffset(tabPositions[selectedTabIndex]),
                                color = GTasksTheme.colors.colorPrimary
                            )
                        },
                        divider = {
                            HorizontalDivider(color = GTasksTheme.colors.borderMuted)
                        }
                    ) {
                        val tabs = listOf("My Work", "Available", "Trades")
                        val counts = listOf(assignedCount, availableCount, pendingTradesCount)
                        tabs.forEachIndexed { index, title ->
                            val count = counts[index]
                            Tab(
                                selected = selectedTabIndex == index,
                                onClick = { selectedTabIndex = index },
                                text = {
                                    Row(
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.Center
                                    ) {
                                        Text(
                                            text = title,
                                            style = MaterialTheme.typography.labelLarge.copy(
                                                fontWeight = if (selectedTabIndex == index) FontWeight.Bold else FontWeight.Medium
                                            ),
                                            color = if (selectedTabIndex == index) GTasksTheme.colors.textPrimary else GTasksTheme.colors.textSecondary
                                        )
                                        
                                        if (count > 0) {
                                            Spacer(modifier = Modifier.width(6.dp))
                                            Box(
                                                modifier = Modifier
                                                    .background(
                                                        color = if (index == 2) {
                                                            GTasksTheme.colors.colorCritical.copy(alpha = 0.15f)
                                                        } else if (selectedTabIndex == index) {
                                                            GTasksTheme.colors.colorPrimary.copy(alpha = 0.15f)
                                                        } else {
                                                            GTasksTheme.colors.borderMuted.copy(alpha = 0.3f)
                                                        },
                                                        shape = RoundedCornerShape(10.dp)
                                                    )
                                                    .border(
                                                        width = 1.dp,
                                                        color = if (index == 2) {
                                                            GTasksTheme.colors.colorCritical.copy(alpha = 0.5f)
                                                        } else if (selectedTabIndex == index) {
                                                            GTasksTheme.colors.colorPrimary.copy(alpha = 0.5f)
                                                        } else {
                                                            GTasksTheme.colors.borderMuted
                                                        },
                                                        shape = RoundedCornerShape(10.dp)
                                                    )
                                                    .padding(horizontal = 6.dp, vertical = 2.dp),
                                                contentAlignment = Alignment.Center
                                            ) {
                                                Text(
                                                    text = count.toString(),
                                                    style = MaterialTheme.typography.labelSmall.copy(
                                                        fontWeight = FontWeight.Bold,
                                                        fontSize = 10.sp
                                                    ),
                                                    color = if (index == 2) {
                                                        GTasksTheme.colors.colorCritical
                                                    } else if (selectedTabIndex == index) {
                                                        GTasksTheme.colors.colorPrimary
                                                    } else {
                                                        GTasksTheme.colors.textSecondary
                                                    }
                                                )
                                            }
                                        }
                                    }
                                }
                            )
                        }
                    }
                }
            }
        },
        floatingActionButton = {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                // Real-time Voice Translation Floating Trigger
                FloatingActionButton(
                    onClick = onTranslateClick,
                    containerColor = GTasksTheme.colors.colorPrimary,
                    contentColor = Color.White,
                    shape = RoundedCornerShape(16.dp),
                    modifier = Modifier.padding(vertical = 8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 16.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Default.Translate,
                            contentDescription = "Translate Icon"
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "Translate",
                            style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold)
                        )
                    }
                }

                // Chat Coach Floating Trigger (Hanna Coach)
                FloatingActionButton(
                    onClick = onChatClick,
                    containerColor = GTasksTheme.colors.colorSecondary,
                    contentColor = Color.White,
                    shape = RoundedCornerShape(16.dp),
                    modifier = Modifier.padding(vertical = 8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 16.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Default.AutoAwesome,
                            contentDescription = "Sparkle Chat Icon"
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "Chat Coach",
                            style = MaterialTheme.typography.labelLarge.copy(fontWeight = FontWeight.Bold)
                        )
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
                            Color(0x1A6366F1), // Translucent Indigo
                            GTasksTheme.colors.bgMain
                        ),
                        radius = 1000f
                    )
                )
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 20.dp)
            ) {
                // 3. Horizontal Scrollable Filter Row (Only applicable to task tabs)
                if (selectedTabIndex < 2) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .horizontalScroll(rememberScrollState()),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Default.FilterList,
                            contentDescription = "Filters Logo",
                            tint = GTasksTheme.colors.textSecondary,
                            modifier = Modifier.size(18.dp)
                        )
                        
                        // Status Filter Chips
                        FilterChip(
                            selected = statusFilter == StatusFilter.ALL_ACTIVE,
                            onClick = { viewModel.setStatusFilter(StatusFilter.ALL_ACTIVE) },
                            label = { Text("All Active", fontSize = 11.sp) }
                        )
                        FilterChip(
                            selected = statusFilter == StatusFilter.PENDING,
                            onClick = { viewModel.setStatusFilter(StatusFilter.PENDING) },
                            label = { Text("Pending", fontSize = 11.sp) }
                        )
                        FilterChip(
                            selected = statusFilter == StatusFilter.IN_PROGRESS,
                            onClick = { viewModel.setStatusFilter(StatusFilter.IN_PROGRESS) },
                            label = { Text("In Progress", fontSize = 11.sp) }
                        )
                        FilterChip(
                            selected = statusFilter == StatusFilter.COMPLETED,
                            onClick = { viewModel.setStatusFilter(StatusFilter.COMPLETED) },
                            label = { Text("Completed", fontSize = 11.sp) }
                        )

                        Spacer(
                            modifier = Modifier
                                .width(1.dp)
                                .height(20.dp)
                                .background(GTasksTheme.colors.borderMuted)
                        )
                        Icon(
                            imageVector = Icons.Default.Sort,
                            contentDescription = "Sort Logo",
                            tint = GTasksTheme.colors.textSecondary,
                            modifier = Modifier.size(18.dp)
                        )

                        // Sort Choice Chips
                        FilterChip(
                            selected = sortBy == SortBy.PRIORITY,
                            onClick = { viewModel.setSortBy(SortBy.PRIORITY) },
                            label = { Text("Priority", fontSize = 11.sp) }
                        )
                        FilterChip(
                            selected = sortBy == SortBy.STATUS,
                            onClick = { viewModel.setSortBy(SortBy.STATUS) },
                            label = { Text("Status", fontSize = 11.sp) }
                        )
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                // 4. Tab Content Router
                when (val state = uiState) {
                    TaskListUiState.Loading -> {
                        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                            CircularProgressIndicator(color = GTasksTheme.colors.colorPrimary)
                        }
                    }
                    is TaskListUiState.Error -> {
                        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Text(state.message, color = GTasksTheme.colors.colorCritical, textAlign = TextAlign.Center)
                                Spacer(modifier = Modifier.height(8.dp))
                                Button(onClick = { viewModel.loadData() }) {
                                    Text("Retry")
                                }
                            }
                        }
                    }
                    is TaskListUiState.Success -> {
                        when (selectedTabIndex) {
                            0 -> MyWorkTab(
                                tasks = state.assignedTasks,
                                onTaskClick = onTaskClick,
                                onTradeClick = { activeTradeTask = it }
                            )
                            1 -> AvailableTab(
                                tasks = state.availableTasks,
                                onClaimClick = { viewModel.claimTask(it.id) },
                                onTaskClick = onTaskClick
                            )
                            2 -> TradesTab(
                                trades = state.pendingTrades,
                                onAccept = { viewModel.acceptTrade(it.id) },
                                onReject = { viewModel.rejectTrade(it.id) }
                            )
                        }
                    }
                }
            }

            // Colleague Trade Selector Dialog
            activeTradeTask?.let { task ->
                TradeProposeDialog(
                    taskName = task.task?.name ?: "Task",
                    colleagues = colleagues,
                    onDismiss = { activeTradeTask = null },
                    onSelectColleague = { colleagueId ->
                        viewModel.proposeTrade(task.id, colleagueId)
                        activeTradeTask = null
                    }
                )
            }
        }
    }
}

// -------------------------------------------------------------------------------------
// Tab 1: My Work (Assigned Tasks)
// -------------------------------------------------------------------------------------
@Composable
private fun MyWorkTab(
    tasks: List<TaskExecution>,
    onTaskClick: (String) -> Unit,
    onTradeClick: (TaskExecution) -> Unit,
    modifier: Modifier = Modifier
) {
    if (tasks.isEmpty()) {
        EmptyQueueView(message = "No tasks assigned to you.\nYou're fully compliant!", icon = Icons.Default.CheckCircle)
    } else {
        LazyColumn(
            verticalArrangement = Arrangement.spacedBy(14.dp),
            contentPadding = PaddingValues(bottom = 80.dp),
            modifier = modifier.fillMaxSize()
        ) {
            items(tasks) { task ->
                TaskCard(
                    task = task,
                    onCardClick = { onTaskClick(task.id) },
                    actionButton = {
                        IconButton(
                            onClick = { onTradeClick(task) },
                            colors = IconButtonDefaults.iconButtonColors(contentColor = GTasksTheme.colors.colorPrimary)
                        ) {
                            Icon(
                                imageVector = Icons.Default.SwapHoriz,
                                contentDescription = "Trade Task",
                                tint = GTasksTheme.colors.colorPrimary,
                                modifier = Modifier.size(22.dp)
                            )
                        }
                    }
                )
            }
        }
    }
}

// -------------------------------------------------------------------------------------
// Tab 2: Available Tasks (Unassigned - "Take Tasks")
// -------------------------------------------------------------------------------------
@Composable
private fun AvailableTab(
    tasks: List<TaskExecution>,
    onClaimClick: (TaskExecution) -> Unit,
    onTaskClick: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    if (tasks.isEmpty()) {
        EmptyQueueView(message = "All tasks have been claimed.\nGreat storefront teamwork!", icon = Icons.Default.CheckCircle)
    } else {
        LazyColumn(
            verticalArrangement = Arrangement.spacedBy(14.dp),
            contentPadding = PaddingValues(bottom = 80.dp),
            modifier = modifier.fillMaxSize()
        ) {
            items(tasks) { task ->
                TaskCard(
                    task = task,
                    onCardClick = { onTaskClick(task.id) },
                    actionButton = {
                        Button(
                            onClick = { onClaimClick(task) },
                            colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorAccent),
                            shape = RoundedCornerShape(8.dp),
                            contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                            modifier = Modifier.height(32.dp)
                        ) {
                            Text("Take", color = Color.White, style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                        }
                    }
                )
            }
        }
    }
}

// -------------------------------------------------------------------------------------
// Tab 3: Trades (Shift & Task Trades)
// -------------------------------------------------------------------------------------
@Composable
private fun TradesTab(
    trades: List<Trade>,
    onAccept: (Trade) -> Unit,
    onReject: (Trade) -> Unit,
    modifier: Modifier = Modifier
) {
    if (trades.isEmpty()) {
        EmptyQueueView(message = "No pending trade offers from peers.", icon = Icons.Default.SwapHoriz)
    } else {
        LazyColumn(
            verticalArrangement = Arrangement.spacedBy(14.dp),
            contentPadding = PaddingValues(bottom = 80.dp),
            modifier = modifier.fillMaxSize()
        ) {
            items(trades) { trade ->
                TradeCard(
                    trade = trade,
                    onAccept = { onAccept(trade) },
                    onReject = { onReject(trade) }
                )
            }
        }
    }
}

// -------------------------------------------------------------------------------------
// Custom Widgets
// -------------------------------------------------------------------------------------

@Composable
private fun TaskCard(
    task: TaskExecution,
    onCardClick: () -> Unit,
    actionButton: @Composable () -> Unit,
    modifier: Modifier = Modifier
) {
    // Priority indicator color
    val indicatorColor = when (task.priority) {
        1 -> GTasksTheme.colors.colorCritical // High (Coral)
        2 -> GTasksTheme.colors.colorPrimary  // Medium (Indigo)
        else -> GTasksTheme.colors.colorAccent // Low (Teal)
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .glassmorphic(shape = RoundedCornerShape(16.dp), elevation = 4.dp)
            .clickable(onClick = onCardClick)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Left Priority Stripe Accent
            Box(
                modifier = Modifier
                    .width(6.dp)
                    .fillMaxHeight()
                    .height(84.dp)
                    .background(indicatorColor)
            )

            // Content
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 14.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = task.task?.name ?: "Unnamed Workload",
                        style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                        color = GTasksTheme.colors.textPrimary,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis
                    )
                    
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        modifier = Modifier.padding(top = 4.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Schedule,
                            contentDescription = "Clock",
                            tint = GTasksTheme.colors.textMuted,
                            modifier = Modifier.size(12.dp)
                        )
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = task.dueAt ?: "On Demand",
                            style = MaterialTheme.typography.labelSmall,
                            color = GTasksTheme.colors.textSecondary
                        )
                        Spacer(modifier = Modifier.width(10.dp))
                        Box(
                            modifier = Modifier
                                .background(GTasksTheme.colors.bgInput, RoundedCornerShape(4.dp))
                                .border(BorderStroke(1.dp, GTasksTheme.colors.borderMuted), RoundedCornerShape(4.dp))
                                .padding(horizontal = 6.dp, vertical = 2.dp)
                        ) {
                            Text(
                                text = task.status ?: "PENDING",
                                style = MaterialTheme.typography.labelSmall.copy(fontSize = 9.sp, fontWeight = FontWeight.Bold),
                                color = if (task.status == "IN_PROGRESS") GTasksTheme.colors.colorAccent else GTasksTheme.colors.textSecondary
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.width(10.dp))
                
                // Card Action (Claim or Trade trigger)
                actionButton()
                
                Spacer(modifier = Modifier.width(4.dp))
                Icon(
                    imageVector = Icons.Default.ChevronRight,
                    contentDescription = "Open Details",
                    tint = GTasksTheme.colors.textMuted
                )
            }
        }
    }
}

@Composable
private fun TradeCard(
    trade: Trade,
    onAccept: () -> Unit,
    onReject: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .glassmorphic(shape = RoundedCornerShape(16.dp), elevation = 4.dp)
            .padding(16.dp)
    ) {
        Column {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
                modifier = Modifier.fillMaxWidth()
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        imageVector = Icons.Default.SwapHoriz,
                        contentDescription = "Trade Icon",
                        tint = GTasksTheme.colors.colorPrimary,
                        modifier = Modifier.size(24.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "TASK COVERAGE OFFER",
                        style = MaterialTheme.typography.labelMedium.copy(fontWeight = FontWeight.Black),
                        color = GTasksTheme.colors.textPrimary
                    )
                }
                
                Box(
                    modifier = Modifier
                        .background(GTasksTheme.colors.colorPrimary.copy(alpha = 0.1f), RoundedCornerShape(4.dp))
                        .padding(horizontal = 6.dp, vertical = 2.dp)
                ) {
                    Text(
                        text = trade.status,
                        style = MaterialTheme.typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                        color = GTasksTheme.colors.colorPrimary
                    )
                }
            }

            Spacer(modifier = Modifier.height(10.dp))
            Text(
                text = "A colleague wants to trade or assign task coverage to you. Accepting this will transfer ownership of the task to your active queue.",
                style = MaterialTheme.typography.bodySmall,
                color = GTasksTheme.colors.textSecondary
            )

            Spacer(modifier = Modifier.height(6.dp))
            Text(
                text = "Task ID: ${trade.taskExecutionId.take(8)}...",
                style = MaterialTheme.typography.labelSmall,
                color = GTasksTheme.colors.textMuted
            )

            Spacer(modifier = Modifier.height(14.dp))
            Row(
                horizontalArrangement = Arrangement.End,
                modifier = Modifier.fillMaxWidth()
            ) {
                // Reject Button
                OutlinedButton(
                    onClick = onReject,
                    colors = ButtonDefaults.outlinedButtonColors(contentColor = GTasksTheme.colors.colorCritical),
                    border = BorderStroke(1.dp, GTasksTheme.colors.colorCritical.copy(alpha = 0.5f)),
                    shape = RoundedCornerShape(8.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                    modifier = Modifier.height(36.dp)
                ) {
                    Icon(Icons.Default.Cancel, contentDescription = null, modifier = Modifier.size(16.dp))
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("Decline", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
                
                Spacer(modifier = Modifier.width(8.dp))
                
                // Accept Button
                Button(
                    onClick = onAccept,
                    colors = ButtonDefaults.buttonColors(containerColor = GTasksTheme.colors.colorAccent),
                    shape = RoundedCornerShape(8.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                    modifier = Modifier.height(36.dp)
                ) {
                    Icon(Icons.Default.CheckCircle, contentDescription = null, modifier = Modifier.size(16.dp), tint = Color.White)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("Accept Cover", color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}

@Composable
private fun EmptyQueueView(
    message: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .padding(top = 80.dp),
        contentAlignment = Alignment.TopCenter
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = GTasksTheme.colors.borderGlow,
                modifier = Modifier.size(48.dp)
            )
            Spacer(modifier = Modifier.height(14.dp))
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = GTasksTheme.colors.textSecondary,
                textAlign = TextAlign.Center,
                lineHeight = 20.sp
            )
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
                    // Option 1: Open Pool Trade (Post to Available)
                    item {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(GTasksTheme.colors.colorPrimary.copy(alpha = 0.1f), RoundedCornerShape(8.dp))
                                .border(BorderStroke(1.dp, GTasksTheme.colors.colorPrimary.copy(alpha = 0.5f)), RoundedCornerShape(8.dp))
                                .clickable { onSelectColleague("") }
                                .padding(horizontal = 14.dp, vertical = 10.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Column {
                                Text(
                                    text = "Post to Available Pool",
                                    style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.Bold),
                                    color = GTasksTheme.colors.colorPrimary
                                )
                                Text(
                                    text = "Open to anyone at this location",
                                    style = MaterialTheme.typography.labelSmall,
                                    color = GTasksTheme.colors.textSecondary
                                )
                            }
                            Icon(
                                imageVector = Icons.Default.AutoAwesome,
                                contentDescription = "Open Trade Sparkle",
                                tint = GTasksTheme.colors.colorPrimary
                            )
                        }
                    }

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
                            Icon(Icons.Default.ChevronRight, contentDescription = null, tint = GTasksTheme.colors.textMuted)
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
