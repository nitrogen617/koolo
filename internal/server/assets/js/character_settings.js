// ==================== RUN CATEGORY FILTERING ====================
const RUN_CATEGORY_MAP = {
    ALL: null, 
    Leveling: ["leveling", "leveling_sequence"],
    ActBoss: ["andariel", "duriel", "mephisto", "diablo", "baal"],
    SUnique: ["pindleskin", "eldritch", "shenk", "threshsocket", "travincal", "fire_eye", "endugu", "countess", "summoner", "nihlathak"],
    A85: ["pit", "ancient_tunnels", "mausoleum", "stony_tomb", "arachnid_lair", "drifter_cavern", "diablo", "baal"],
    Key: ["countess", "summoner", "nihlathak"],
    Special: ["cows", "tristram", "terror_zone", "quests", "lower_kurast_chest"],
    Utils: ["mule", "utility", "shopping", "development"],
};

function filterRunsByCategory(category) {
    const runList = document.querySelectorAll('#disabled_runs li');
    if (!runList) return;
    const allowed = RUN_CATEGORY_MAP[category];
    runList.forEach(li => {
        if (!allowed || allowed.includes(li.getAttribute('value'))) {
            li.style.display = '';
        } else {
            li.style.display = 'none';
        }
    });
}

function setupRunCategoryTabs() {
    const tabs = document.querySelectorAll('.run-category-tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', function() {
            tabs.forEach(t => t.classList.remove('active'));
            this.classList.add('active');
            // Keep the selected tab visually highlighted
            tabs.forEach(t => {
                if (t.classList.contains('active')) {
                    t.style.background = '#161a23';
                    t.style.color = '#fff';
                } else {
                    t.style.background = 'none';
                    t.style.color = '#161a23';
                }
            });
            const category = this.getAttribute('data-category');
            filterRunsByCategory(category);
        });
    });
    // Apply initial tab styling on load
    tabs.forEach(t => {
        if (t.classList.contains('active')) {
            t.style.background = '#161a23';
            t.style.color = '#fff';
        } else {
            t.style.background = 'none';
            t.style.color = '#161a23';
        }
    });
}

// Toggle game name/password fields based on Game/Companion checkboxes
function updateGameNamePasswordVisibility() {
    // Query fresh DOM nodes each time
    var createLobbyGames = document.querySelector('input[name="createLobbyGames"]');
    var companionLeader = document.querySelector('input[name="companionLeader"]');
    var gameNamePasswordFields = document.getElementById('game-name-password-fields');
    if (!createLobbyGames || !companionLeader || !gameNamePasswordFields) return;
    if (createLobbyGames.checked || companionLeader.checked) {
        gameNamePasswordFields.style.display = '';
    } else {
        gameNamePasswordFields.style.display = 'none';
    }
}
window.updateGameNamePasswordVisibility = updateGameNamePasswordVisibility;

function attachGameNamePasswordListeners() {
    var createLobbyGames = document.querySelector('input[name="createLobbyGames"]');
    var companionLeader = document.querySelector('input[name="companionLeader"]');
    if (createLobbyGames) {
        createLobbyGames.removeEventListener('change', updateGameNamePasswordVisibility);
        createLobbyGames.addEventListener('change', updateGameNamePasswordVisibility);
    }
    if (companionLeader) {
        companionLeader.removeEventListener('change', updateGameNamePasswordVisibility);
        companionLeader.addEventListener('change', updateGameNamePasswordVisibility);
    }
}

function ensureGameNamePasswordVisibility() {
    attachGameNamePasswordListeners();
    updateGameNamePasswordVisibility();
}

// Warning: packet usage toggle
function initializePacketWarningToggle() {
    const warningCard = document.querySelector('[data-collapsible="packet-warning"]');
    if (!warningCard) return;

    const toggleButton = warningCard.querySelector('.warning-toggle');
    const warningBody = warningCard.querySelector('.warning-body');
    if (!toggleButton || !warningBody) return;

    const checkboxes = Array.from(warningBody.querySelectorAll('input[type="checkbox"]'));

    function setExpanded(expanded) {
        warningCard.classList.toggle('open', expanded);
        warningBody.style.display = expanded ? '' : 'none';
        toggleButton.setAttribute('aria-expanded', expanded ? 'true' : 'false');
        toggleButton.textContent = expanded ? 'HIDE OPTIONS' : 'SHOW OPTIONS';
    }

    // Auto-open if any option is already enabled
    const startOpen = checkboxes.some(cb => cb.checked);
    setExpanded(startOpen);

    toggleButton.addEventListener('click', function () {
        const isOpen = warningCard.classList.contains('open');
        setExpanded(!isOpen);
    });

    // If a checkbox is enabled while closed, open the panel so the change is visible
    checkboxes.forEach(cb => {
        cb.addEventListener('change', function () {
            if (cb.checked) setExpanded(true);
        });
    });
}

document.addEventListener('DOMContentLoaded', function () {
    // Enable drag handles for run-category mini tabs (Sortable.js)
    var runCategoryTabs = document.getElementById('run-category-tabs');
    if (runCategoryTabs && typeof Sortable !== 'undefined') {
        runCategoryTabs.querySelectorAll('.run-category-tab').forEach(function (tab) {
            if (!tab.querySelector('.run-category-handle')) {
                const handle = document.createElement('span');
                handle.className = 'run-category-handle';
                handle.innerHTML = '\u22ee';
                tab.insertBefore(handle, tab.firstChild);
            }
        });
        new Sortable(runCategoryTabs, {
            animation: 150,
            ghostClass: 'sortable-ghost',
            chosenClass: 'sortable-chosen',
            dragClass: 'sortable-drag',
            handle: '.run-category-handle',
            filter: ".run-category-tab[data-category='ALL']",
            preventOnFilter: false,
            onMove: function (evt) {
                // Prevent dragging the ALL tab
                return !evt.related.classList.contains('run-category-tab') || evt.related.dataset.category !== 'ALL';
            }
        });
    }
    setupRunCategoryTabs();
    initializePacketWarningToggle();
    // Make sure game name/password visibility is correct on first paint
    setTimeout(ensureGameNamePasswordVisibility, 0);
    setTimeout(ensureGameNamePasswordVisibility, 100);
    setTimeout(ensureGameNamePasswordVisibility, 300);
    // Default to ALL filter on load
    filterRunsByCategory('ALL');
});
// ==================== TAB SYSTEM ====================
const TAB_ORDER_STORAGE_KEY = 'koolo:tabOrder';
const ACTIVE_TAB_STORAGE_KEY = 'koolo:activeTab';

// Initialize tabs on page load
window.onload = function () {
    initializeTabs();
    initializeRunLists();
    initializeOtherFeatures();
    organizeRecipes();
    setTimeout(ensureGameNamePasswordVisibility, 0);
    setTimeout(ensureGameNamePasswordVisibility, 100);
    setTimeout(ensureGameNamePasswordVisibility, 300);
    initializeCardDrag();
}

// Cube recipes toggle (separate listener to avoid layout jump)
document.addEventListener('DOMContentLoaded', function () {
    var cubeToggle = document.querySelector('input[name="enableCubeRecipes"]');
    var cubeTargets = document.querySelectorAll('.cube-toggle-target');
    function updateCubeVisibility() {
        if (!cubeToggle || !cubeTargets.length) return;
        var show = cubeToggle.checked;
        cubeTargets.forEach(function (el) {
            if (show) {
                el.classList.remove('cube-hidden');
            } else {
                el.classList.add('cube-hidden');
            }
            var inputs = el.querySelectorAll('input, select, textarea, button');
            inputs.forEach(function (inp) {
                inp.disabled = !show;
            });
        });
    }
    if (cubeToggle) {
        cubeToggle.addEventListener('change', updateCubeVisibility);
        updateCubeVisibility();
    }

    // Clone supervisor redirect for instant prefilling
    var cloneSelect = document.getElementById('cloneFrom');
    if (cloneSelect) {
        cloneSelect.addEventListener('change', function () {
            var val = cloneSelect.value;
            if (val) {
                window.location = '/supervisorSettings?clone=' + encodeURIComponent(val);
            } else {
                window.location = '/supervisorSettings';
            }
        });
    }
});

// Card drag & drop per tab
function initializeCardDrag() {
    if (typeof Sortable === 'undefined') return;
    const tabContents = document.querySelectorAll('.tab-content');

    tabContents.forEach(function (tab) {
        const tabId = tab.id || tab.dataset.tab || 'default';
        const storageKey = `koolo:cardOrder:${tabId}`;
        const cards = Array.from(tab.querySelectorAll('.section-card'));
        if (!cards.length) return;

        // Insert drag handle if missing
        cards.forEach(function (card) {
            if (!card.querySelector('.card-handle')) {
                const handle = document.createElement('div');
                handle.className = 'card-handle';
                handle.innerHTML = '\u22ee';
                const heading = card.querySelector('.section-heading h3, .section-heading h4, .section-heading h5, h3, h4, h5, h6');
                if (heading) {
                    card.insertBefore(handle, card.firstChild);
                } else {
                    card.insertBefore(handle, card.firstChild);
                }
            }
        });

        // Apply saved order
        try {
            const savedOrder = JSON.parse(localStorage.getItem(storageKey) || '[]');
            savedOrder.forEach(function (id) {
                const card = tab.querySelector(`.section-card[data-card-id="${id}"]`);
                if (card) tab.appendChild(card);
            });
        } catch (e) {
            console.error('Failed to load card order', e);
        }

        new Sortable(tab, {
            animation: 150,
            handle: '.card-handle',
            draggable: '.section-card',
            onEnd: function () {
                const order = Array.from(tab.querySelectorAll('.section-card'))
                    .map(function (c) { return c.dataset.cardId; })
                    .filter(Boolean);
                try {
                    localStorage.setItem(storageKey, JSON.stringify(order));
                } catch (e) {
                    console.error('Failed to save card order', e);
                }
            }
        });
    });
}

// Cube recipes visibility toggle
document.addEventListener('DOMContentLoaded', function () {
    var cubeToggle = document.querySelector('input[name="enableCubeRecipes"]');
    var cubeBody = document.getElementById('cube-recipes-body');
    var cubeGrid = document.getElementById('cube-recipes-grid');

    function updateCubeVisibility() {
        if (!cubeToggle || !cubeBody || !cubeGrid) return;
        var show = cubeToggle.checked;
        cubeBody.style.display = show ? '' : 'none';
        cubeGrid.style.display = show ? '' : 'none';
    }

    if (cubeToggle) {
        cubeToggle.addEventListener('change', updateCubeVisibility);
        updateCubeVisibility();
    }
});

// ==================== TAB INITIALIZATION ====================
function initializeTabs() {
    const tabNavigation = document.getElementById('tabs-navigation');
    const tabItems = document.querySelectorAll('.tab-item');
    const tabContents = document.querySelectorAll('.tab-content');

    // Insert drag handles on tabs
    tabItems.forEach(tab => {
        if (!tab.querySelector('.tab-drag-handle')) {
            const handle = document.createElement('span');
            handle.className = 'tab-drag-handle';
            handle.innerHTML = '\u22ee';
            tab.insertBefore(handle, tab.firstChild);
        }
    });

    // Load saved tab order from localStorage
    loadTabOrder();

    // Make tabs sortable (drag & drop)
    new Sortable(tabNavigation, {
        animation: 150,
        handle: '.tab-drag-handle',
        onEnd: function (evt) {
            saveTabOrder();
        }
    });

    // Add click event listeners to tabs
    tabItems.forEach(tab => {
        tab.addEventListener('click', function () {
            switchTab(this.getAttribute('data-tab'));
        });
    });

    // Load last active tab or default to first tab
    const lastActiveTab = localStorage.getItem(ACTIVE_TAB_STORAGE_KEY);
    if (lastActiveTab && document.querySelector(`[data-tab="${lastActiveTab}"]`)) {
        switchTab(lastActiveTab);
    } else {
        const firstTab = tabItems[0]?.getAttribute('data-tab');
        if (firstTab) switchTab(firstTab);
    }
}

// ==================== TAB SWITCHING ====================
function switchTab(tabName) {
    const tabItems = document.querySelectorAll('.tab-item');
    const tabContents = document.querySelectorAll('.tab-content');

    // Remove active class from all tabs and contents
    tabItems.forEach(tab => tab.classList.remove('active'));
    tabContents.forEach(content => content.classList.remove('active'));

    // Add active class to selected tab and content
    const selectedTab = document.querySelector(`[data-tab="${tabName}"]`);
    const selectedContent = document.getElementById(`tab-${tabName}`);

    if (selectedTab && selectedContent) {
        selectedTab.classList.add('active');
        selectedContent.classList.add('active');

        // Save active tab to localStorage
        localStorage.setItem(ACTIVE_TAB_STORAGE_KEY, tabName);
    }
    // Recheck game name/password visibility in case the tab change moved fields
    setTimeout(ensureGameNamePasswordVisibility, 0);
    setTimeout(ensureGameNamePasswordVisibility, 100);
    setTimeout(ensureGameNamePasswordVisibility, 300);
}

// ==================== TAB ORDER MANAGEMENT ====================
function saveTabOrder() {
    const tabNavigation = document.getElementById('tabs-navigation');
    const tabItems = tabNavigation.querySelectorAll('.tab-item');
    const tabOrder = Array.from(tabItems).map(tab => tab.getAttribute('data-tab'));
    
    try {
        localStorage.setItem(TAB_ORDER_STORAGE_KEY, JSON.stringify(tabOrder));
    } catch (error) {
        console.error('Failed to save tab order:', error);
    }
}

function loadTabOrder() {
    try {
        const savedOrder = localStorage.getItem(TAB_ORDER_STORAGE_KEY);
        if (!savedOrder) return;

        const tabOrder = JSON.parse(savedOrder);
        const tabNavigation = document.getElementById('tabs-navigation');
        const tabItems = Array.from(tabNavigation.querySelectorAll('.tab-item'));

        // Reorder tabs based on saved order
        tabOrder.forEach(tabName => {
            const tab = tabItems.find(item => item.getAttribute('data-tab') === tabName);
            if (tab) {
                tabNavigation.appendChild(tab);
            }
        });
    } catch (error) {
        console.error('Failed to load tab order:', error);
    }
}

// ==================== RUN LISTS INITIALIZATION ====================
function initializeRunLists() {
    let enabled_runs_ul = document.getElementById('enabled_runs');
    let disabled_runs_ul = document.getElementById('disabled_runs');
    let searchInput = document.getElementById('search-disabled-runs');

    new Sortable(enabled_runs_ul, {
        group: 'runs',
        animation: 150,
        onSort: function (evt) {
            updateEnabledRunsHiddenField();
        },
        onAdd: function (evt) {
            updateButtonForEnabledRun(evt.item);
        }
    });

    new Sortable(disabled_runs_ul, {
        group: 'runs',
        animation: 150,
        onAdd: function (evt) {
            updateButtonForDisabledRun(evt.item);
        }
    });

    searchInput.addEventListener('input', function () {
        filterDisabledRuns(searchInput.value);
    });

    // Add event listeners for add and remove buttons
    document.addEventListener('click', function (e) {
        if (e.target.closest('.remove-run')) {
            e.preventDefault();
            const runElement = e.target.closest('li');
            moveRunToDisabled(runElement);
        } else if (e.target.closest('.add-run')) {
            e.preventDefault();
            const runElement = e.target.closest('li');
            moveRunToEnabled(runElement);
        }
    });

    updateEnabledRunsHiddenField();

    const buildSelectElement = document.querySelector('select[name="characterClass"]');
    buildSelectElement.addEventListener('change', function () {
        const selectedBuild = buildSelectElement.value;
        const levelingBuilds = ['paladin', 'sorceress_leveling', 'druid_leveling', 'amazon_leveling', 'necromancer', 'assassin'];

        const enabledRunListElement = document.getElementById('enabled_runs');
        if (!enabledRunListElement) return;

        const enabledRuns = Array.from(enabledRunListElement.querySelectorAll('li')).map(li => li.getAttribute('value'));
        const isLevelingRunEnabled = enabledRuns.includes('leveling') || enabledRuns.includes('leveling_sequence');
        const hasOtherRunsEnabled = enabledRuns.length > 1;

        if (levelingBuilds.includes(selectedBuild) && (!isLevelingRunEnabled || hasOtherRunsEnabled)) {
            alert("This profile requires enabling the leveling run. Please add only the 'leveling' run to the enabled run list and remove the others.");
        }
    });
}

// ==================== RUN LIST FUNCTIONS ====================
function updateEnabledRunsHiddenField() {
    let listItems = document.querySelectorAll('#enabled_runs li');
    let values = Array.from(listItems).map(function (item) {
        return item.getAttribute("value");
    });
    document.getElementById('gameRuns').value = JSON.stringify(values);
}

function filterDisabledRuns(searchTerm) {
    let listItems = document.querySelectorAll('#disabled_runs li');
    searchTerm = searchTerm.toLowerCase();
    listItems.forEach(function (item) {
        let runName = item.getAttribute("value").toLowerCase();
        if (runName.includes(searchTerm)) {
            item.style.display = '';
        } else {
            item.style.display = 'none';
        }
    });
}

function checkLevelingProfile() {
    const levelingProfiles = [
        "sorceress_leveling",
        "paladin",
        "druid_leveling",
        "amazon_leveling",
        "necromancer",
        "assassin"
    ];

    const characterClass = document.getElementById('characterClass').value;

    if (levelingProfiles.includes(characterClass)) {
        const confirmation = confirm("This profile requires the leveling run profile, would you like to clear enabled run profiles and select the leveling profile?");
        if (confirmation) {
            clearEnabledRuns();
            selectLevelingProfile();
        }
    }
}

function moveRunToDisabled(runElement) {
    const disabledRunsUl = document.getElementById('disabled_runs');
    updateButtonForDisabledRun(runElement);
    disabledRunsUl.appendChild(runElement);
    updateEnabledRunsHiddenField();
}

function moveRunToEnabled(runElement) {
    const enabledRunsUl = document.getElementById('enabled_runs');
    updateButtonForEnabledRun(runElement);
    enabledRunsUl.appendChild(runElement);
    updateEnabledRunsHiddenField();
}

function updateButtonForEnabledRun(runElement) {
    const button = runElement.querySelector('button');
    button.classList.remove('add-run');
    button.classList.add('remove-run');
    button.title = "Remove run";
    button.innerHTML = '<i class="bi bi-dash"></i>';
}

function updateButtonForDisabledRun(runElement) {
    const button = runElement.querySelector('button');
    button.classList.remove('remove-run');
    button.classList.add('add-run');
    button.title = "Add run";
    button.innerHTML = '<i class="bi bi-plus"></i>';
}

// ==================== OTHER FEATURES INITIALIZATION ====================
function initializeOtherFeatures() {
    const schedulerEnabled = document.querySelector('input[name="schedulerEnabled"]');
    const schedulerSettings = document.getElementById('scheduler-settings');
    const characterClassSelect = document.querySelector('select[name="characterClass"]');
    const berserkerBarbOptions = document.querySelector('.berserker-barb-options');
    const novaSorceressOptions = document.querySelector('.nova-sorceress-options');
    const bossStaticThresholdInput = document.getElementById('novaBossStaticThreshold');
    const mosaicAssassinOptions = document.querySelector('.mosaic-assassin-options');
    const blizzardSorceressOptions = document.querySelector('.blizzard-sorceress-options');
    const sorceressLevelingOptions = document.querySelector('.sorceress_leveling-options');
    const runewordSearchInput = document.getElementById('search-runewords');
    const useTeleportCheckbox = document.getElementById('characterUseTeleport');
    const useExtraBuffsCheckbox = document.getElementById('characterUseExtraBuffs');
    const clearPathDistContainer = document.getElementById('clearPathDistContainer');
    const useExtraBuffsDistContainer = document.getElementById('useExtraBuffsDistContainer');
    const clearPathDistInput = document.getElementById('clearPathDist');
    const clearPathDistValue = document.getElementById('clearPathDistValue');

    if (bossStaticThresholdInput) {
        bossStaticThresholdInput.addEventListener('input', handleBossStaticThresholdChange);
    }

    function toggleSchedulerVisibility() {
        schedulerSettings.style.display = schedulerEnabled.checked ? 'grid' : 'none';
    }

    function updateCharacterOptions() {
        const selectedClass = characterClassSelect.value;
        const noSettingsMessage = document.getElementById('no-settings-message');
        const berserkerBarbOptions = document.querySelector('.berserker-barb-options');
        const novaSorceressOptions = document.querySelector('.nova-sorceress-options');
        const mosaicAssassinOptions = document.querySelector('.mosaic-assassin-options');
        const blizzardSorceressOptions = document.querySelector('.blizzard-sorceress-options');
        const sorceressLevelingOptions = document.querySelector('.sorceress_leveling-options');
        // Hide all options first
        berserkerBarbOptions.style.display = 'none';
        novaSorceressOptions.style.display = 'none';
        mosaicAssassinOptions.style.display = 'none';
        blizzardSorceressOptions.style.display = 'none';
        sorceressLevelingOptions.style.display = 'none';
        noSettingsMessage.style.display = 'none';

        // Show relevant options based on class
        if (selectedClass === 'berserker') {
            berserkerBarbOptions.style.display = 'block';
        } else if (selectedClass === 'nova' || selectedClass === 'lightsorc') {
            novaSorceressOptions.style.display = 'block';
            updateNovaSorceressOptions();
        } else if (selectedClass === 'mosaic') {
            mosaicAssassinOptions.style.display = 'block';
        } else if (selectedClass === 'sorceress') {
            blizzardSorceressOptions.style.display = 'block';
        } else if (selectedClass === 'sorceress_leveling') {
            sorceressLevelingOptions.style.display = 'block';
        } else {
            noSettingsMessage.style.display = 'block';
        }
    }
    function toggleClearPathVisibility() {
        if (useTeleportCheckbox && clearPathDistContainer) {
            if (useTeleportCheckbox.checked) {
                clearPathDistContainer.style.display = 'none';
            } else {
                clearPathDistContainer.style.display = 'block';
            }
        }
    }
    function toggleUseExtraBuffsVisibility() {
        if (useExtraBuffsCheckbox && useExtraBuffsDistContainer) {
            if (useExtraBuffsCheckbox.checked) {
                useExtraBuffsDistContainer.style.display = 'block';
            } else {
                useExtraBuffsDistContainer.style.display = 'none';
            }
        }
    }

    // Update the displayed value when the slider changes
    function updateClearPathValue() {
        if (clearPathDistInput && clearPathDistValue) {
            clearPathDistValue.textContent = clearPathDistInput.value;

            // Calculate tooltip position based on slider value
            const min = parseFloat(clearPathDistInput.min);
            const max = parseFloat(clearPathDistInput.max);
            const value = parseFloat(clearPathDistInput.value);
            const percentage = ((value - min) / (max - min)) * 100;

            // Position the tooltip above the thumb
            clearPathDistValue.style.left = `calc(${percentage}% + (${8 - percentage * 0.15}px))`;
        }
    }

    // Show/hide tooltip on mouse interaction
    function showClearPathTooltip() {
        if (clearPathDistValue) {
            clearPathDistValue.style.opacity = '1';
            clearPathDistValue.style.pointerEvents = 'none';
        }
    }

    function hideClearPathTooltip() {
        if (clearPathDistValue) {
            clearPathDistValue.style.opacity = '0';
        }
    }

    // Set up event listeners
    if (useTeleportCheckbox) {
        useTeleportCheckbox.addEventListener('change', toggleClearPathVisibility);
        // Initialize visibility
        toggleClearPathVisibility();
    }

    // Set up event listeners
    if (useExtraBuffsCheckbox) {
        useExtraBuffsCheckbox.addEventListener('change', toggleUseExtraBuffsVisibility);
        // Initialize visibility
        toggleUseExtraBuffsVisibility();
    }

    if (clearPathDistInput) {
        clearPathDistInput.addEventListener('input', updateClearPathValue);
        clearPathDistInput.addEventListener('mousedown', showClearPathTooltip);
        clearPathDistInput.addEventListener('mouseup', hideClearPathTooltip);
        clearPathDistInput.addEventListener('mouseleave', hideClearPathTooltip);
        // Initialize value display and hide tooltip
        updateClearPathValue();
        hideClearPathTooltip();
    }

    function updateNovaSorceressOptions() {
        const selectedDifficulty = document.getElementById('gameDifficulty').value;
        updateBossStaticThresholdMin(selectedDifficulty);
        handleBossStaticThresholdChange();
    }

    function updateBossStaticThresholdMin(difficulty) {
        const input = document.getElementById('novaBossStaticThreshold');
        let minValue;
        switch (difficulty) {
            case 'normal':
                minValue = 1;
                break;
            case 'nightmare':
                minValue = 33;
                break;
            case 'hell':
                minValue = 50;
                break;
            default:
                minValue = 65;
        }
        input.min = minValue;

        // Ensure the current value is not less than the new minimum
        if (parseInt(input.value) < minValue) {
            input.value = minValue;
        }
    }

    characterClassSelect.addEventListener('change', updateCharacterOptions);
    document.getElementById('gameDifficulty').addEventListener('change', function () {
        if (characterClassSelect.value === 'nova' || characterClassSelect.value === 'lightsorc') {
            updateNovaSorceressOptions();
        }
    });

    characterClassSelect.addEventListener('change', updateCharacterOptions);
    updateCharacterOptions(); // Call this initially to set the correct state

    // Set initial state
    toggleSchedulerVisibility();
    updateNovaSorceressOptions();

    schedulerEnabled.addEventListener('change', toggleSchedulerVisibility);

    document.querySelectorAll('.add-time-range').forEach(button => {
        button.addEventListener('click', function () {
            const day = this.dataset.day;
            const timeRangesDiv = this.previousElementSibling;
            if (timeRangesDiv) {
                const newTimeRange = document.createElement('div');
                newTimeRange.className = 'time-range';
                newTimeRange.innerHTML = `
                    <input type="time" name="scheduler[${day}][start][]" required>
                    <span>to</span>
                    <input type="time" name="scheduler[${day}][end][]" required>
                    <button type="button" class="remove-time-range"><i class="bi bi-trash"></i></button>
                `;
                timeRangesDiv.appendChild(newTimeRange);
            }
        });
    });

    document.addEventListener('click', function (e) {
        if (e.target.closest('.remove-time-range')) {
            e.target.closest('.time-range').remove();
        }
    });

    document.getElementById('tzTrackAll')?.addEventListener('change', function (e) {
        document.querySelectorAll('.tzTrackCheckbox').forEach(checkbox => {
            checkbox.checked = e.target.checked;
        });
    });

    function filterRunewords(searchTerm = '') { // Default parameter to ensure previously checked runewords show before searching
        let listItems = document.querySelectorAll('.runeword-item');
        searchTerm = searchTerm.toLowerCase();
        const showAllToggle = document.getElementById('runeword-show-all');
        const forceShowAll = showAllToggle ? showAllToggle.checked : false;

        listItems.forEach(function (item) {
            const isChecked = item.querySelector('input[type="checkbox"]').checked;
            const rwName = item.querySelector('.runeword-name').textContent.toLowerCase();

            if (forceShowAll || isChecked || (searchTerm && rwName.includes(searchTerm))) {
                item.style.display = '';
            } else {
                item.style.display = 'none';
            }
        });
    }

    if (runewordSearchInput) {
        runewordSearchInput.addEventListener('input', function () {
            filterRunewords(runewordSearchInput.value);
        });

        document.addEventListener('change', function (e) {
            if (e.target.matches('.runeword-item input[type="checkbox"]')) {
                filterRunewords(runewordSearchInput.value);
            }
            if (e.target.id === 'runeword-show-all') {
                filterRunewords(runewordSearchInput.value);
            }
        });

        filterRunewords();
    }

    const levelingSequenceSelect = document.getElementById('gameLevelingSequenceSelect');
    const levelingSequenceAddBtn = document.getElementById('levelingSequenceAddBtn');
    const levelingSequenceEditBtn = document.getElementById('levelingSequenceEditBtn');
    const levelingSequenceDeleteBtn = document.getElementById('levelingSequenceDeleteBtn');
    const LAST_SEQUENCE_KEY = 'koolo:lastSequenceName';
    const REFRESH_FLAG_KEY = 'koolo:sequenceRefreshRequired';
    const sequenceFilesEndpoint = '/api/sequence-editor/files';
    const sequenceDeleteEndpoint = '/api/sequence-editor/delete';

    const updateLevelingSequenceActionState = () => {
        const hasSelection = Boolean(levelingSequenceSelect && levelingSequenceSelect.value);
        if (levelingSequenceEditBtn) {
            levelingSequenceEditBtn.disabled = !hasSelection;
        }
        if (levelingSequenceDeleteBtn) {
            levelingSequenceDeleteBtn.disabled = !hasSelection;
        }
    };

    const rebuildLevelingSequenceOptions = (files, desiredSelection) => {
        if (!levelingSequenceSelect) {
            return;
        }

        const fragment = document.createDocumentFragment();
        const placeholder = document.createElement('option');
        placeholder.value = '';
        placeholder.disabled = true;
        placeholder.textContent = 'Select a sequence file';
        if (!desiredSelection) {
            placeholder.selected = true;
        }
        fragment.appendChild(placeholder);

        const hasDesired = desiredSelection && files.includes(desiredSelection);

        if (desiredSelection && !hasDesired) {
            const missingOption = document.createElement('option');
            missingOption.value = desiredSelection;
            missingOption.textContent = `${desiredSelection} (missing)`;
            missingOption.selected = true;
            fragment.appendChild(missingOption);
        }

        files.forEach((fileName) => {
            const option = document.createElement('option');
            option.value = fileName;
            option.textContent = fileName;
            if (fileName === desiredSelection) {
                option.selected = true;
            }
            fragment.appendChild(option);
        });

        levelingSequenceSelect.innerHTML = '';
        levelingSequenceSelect.appendChild(fragment);

        if (desiredSelection && hasDesired) {
            levelingSequenceSelect.value = desiredSelection;
        }
    };

    const refreshLevelingSequenceOptions = async (preferredSelection) => {
        if (!levelingSequenceSelect) {
            return false;
        }

        const targetSelection = typeof preferredSelection === 'string' ? preferredSelection : levelingSequenceSelect.value;

        try {
            const response = await fetch(sequenceFilesEndpoint, {
                headers: { 'Accept': 'application/json' },
            });
            if (!response.ok) {
                throw new Error(`Failed to fetch sequence files (${response.status})`);
            }
            const payload = await response.json();
            const files = Array.isArray(payload.files) ? payload.files : [];
            rebuildLevelingSequenceOptions(files, targetSelection);
            updateLevelingSequenceActionState();
            return true;
        } catch (error) {
            console.error('Unable to refresh leveling sequence list', error);
            return false;
        }
    };

    const maybeRefreshSequencesFromStorage = async () => {
        if (!levelingSequenceSelect || !window.localStorage) {
            return;
        }

        let refreshFlag;
        try {
            refreshFlag = window.localStorage.getItem(REFRESH_FLAG_KEY);
        } catch (error) {
            console.warn('Unable to read sequence refresh flag', error);
            return;
        }

        if (!refreshFlag) {
            return;
        }

        let desiredSelection = '';
        try {
            desiredSelection = window.localStorage.getItem(LAST_SEQUENCE_KEY) || '';
        } catch (error) {
            console.warn('Unable to read last sequence name', error);
        }

        const refreshed = await refreshLevelingSequenceOptions(desiredSelection);
        if (refreshed) {
            try {
                window.localStorage.removeItem(REFRESH_FLAG_KEY);
                if (desiredSelection) {
                    window.localStorage.removeItem(LAST_SEQUENCE_KEY);
                }
            } catch (error) {
                console.warn('Unable to clear sequence refresh flags', error);
            }
        }
    };

    if (levelingSequenceSelect) {
        levelingSequenceSelect.addEventListener('change', updateLevelingSequenceActionState);
    }
    if (levelingSequenceDeleteBtn) {
        levelingSequenceDeleteBtn.addEventListener('click', async () => {
            if (!levelingSequenceSelect || !levelingSequenceSelect.value) {
                return;
            }

            const targetName = levelingSequenceSelect.value;
            const confirmed = window.confirm(`Delete "${targetName}"? This cannot be undone.`);
            if (!confirmed) {
                return;
            }

            try {
                const response = await fetch(sequenceDeleteEndpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Accept': 'application/json',
                    },
                    body: JSON.stringify({ name: targetName }),
                });

                if (!response.ok) {
                    const message = await response.text();
                    throw new Error(message || `Failed to delete sequence (${response.status})`);
                }

                await refreshLevelingSequenceOptions('');
                updateLevelingSequenceActionState();
            } catch (error) {
                console.error('Failed to delete leveling sequence', error);
                alert('Unable to delete the selected sequence. Please check the logs for more information.');
            }
        });
    }


    if (levelingSequenceAddBtn) {
        levelingSequenceAddBtn.addEventListener('click', () => {
            window.open('/sequence-editor', '_blank');
        });
    }

    if (levelingSequenceEditBtn) {
        levelingSequenceEditBtn.addEventListener('click', () => {
            if (!levelingSequenceSelect || !levelingSequenceSelect.value) {
                return;
            }
            const encoded = encodeURIComponent(levelingSequenceSelect.value);
            window.open(`/sequence-editor?sequence=${encoded}`, '_blank');
        });
    }

    window.addEventListener('focus', () => {
        void maybeRefreshSequencesFromStorage();
    });

    document.addEventListener('visibilitychange', () => {
        if (!document.hidden) {
            void maybeRefreshSequencesFromStorage();
        }
    });

    updateLevelingSequenceActionState();
}

function handleBossStaticThresholdChange() {
    const input = document.getElementById('novaBossStaticThreshold');
    const selectedDifficulty = document.getElementById('gameDifficulty').value;
    let minValue;
    switch (selectedDifficulty) {
        case 'normal':
            minValue = 1;
            break;
        case 'nightmare':
            minValue = 33;
            break;
        case 'hell':
            minValue = 50;
            break;
        default:
            minValue = 65;
    }

    let value = parseInt(input.value);
    if (isNaN(value) || value < minValue) {
        value = minValue;
    } else if (value > 100) {
        value = 100;
    }
    input.value = value;
}
// ==================== CUBE RECIPES ORGANIZER (No Select All) ====================
function organizeRecipes() {
    const source = document.getElementById('hidden-recipe-source');
    const navContainer = document.getElementById('recipe-tab-nav');
    const gridContainer = document.getElementById('recipe-grid-container');
    
    // toggleBtn 관련 변수 및 로직 삭제됨

    if (!source || !navContainer || !gridContainer) return;

    const allItems = Array.from(source.querySelectorAll('.recipe-item'));

    // 카테고리 정의
    const categories = [
        { id: 'all', title: 'All', items: allItems, keywords: [] },
        { id: 'active', title: 'Active', items: [], keywords: [] },
        { id: 'runes', title: 'Runes', items: [], keywords: ['Upgrade'] },
        { id: 'crafting', title: 'Crafting', items: [], keywords: ['Caster', 'Blood', 'Safety', 'Hitpower'] },
        { id: 'sockets', title: 'Sockets', items: [], keywords: ['Add Sockets'] },
        { id: 'gems', title: 'Gems', items: [], keywords: ['Perfect'] },
        { id: 'misc', title: 'Misc', items: [], keywords: [] }
    ];

    // 아이템 분류 로직
    allItems.forEach(item => {
        const name = item.getAttribute('data-name');
        let placed = false;

        for (let i = 2; i < categories.length - 1; i++) {
            const cat = categories[i];
            if (cat.keywords.some(k => name.includes(k))) {
                cat.items.push(item);
                placed = true;
                break;
            }
        }
        if (!placed) {
            categories[categories.length - 1].items.push(item);
        }
    });

    // 탭 버튼 생성
    navContainer.innerHTML = ''; 
    let activeTabId = 'all';

    categories.forEach(cat => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'recipe-tab-btn';
        btn.dataset.id = cat.id;
        
        updateButtonText(btn, cat);

        if (cat.id === 'all') btn.classList.add('active');

        btn.addEventListener('click', () => {
            document.querySelectorAll('.recipe-tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            activeTabId = cat.id;
            
            if (cat.id === 'active') {
                const activeItems = allItems.filter(item => item.querySelector('input').checked);
                renderGrid(activeItems);
            } else {
                renderGrid(cat.items);
            }
        });

        navContainer.appendChild(btn);
    });

    // 그리드 렌더링 함수
    function renderGrid(items) {
        gridContainer.innerHTML = '';
        if (items.length === 0) {
            gridContainer.innerHTML = '<div style="color:gray; padding:20px; text-align:center; grid-column: 1/-1;">No items found.</div>';
            return;
        }
        items.forEach(item => gridContainer.appendChild(item));
    }

    // 버튼 텍스트 업데이트 (Active 카운트용)
    function updateButtonText(btn, cat) {
        if (cat.id === 'active') {
            const checkedCount = allItems.filter(i => i.querySelector('input').checked).length;
            btn.textContent = `${cat.title} (${checkedCount})`;
        } else {
            btn.textContent = `${cat.title} (${cat.items.length})`;
        }
    }

    // 체크박스 변경 감지 (Active 탭 숫자 업데이트)
    gridContainer.addEventListener('change', (e) => {
        if (e.target.matches('input[type="checkbox"]')) {
            refreshActiveCount();
        }
    });

    function refreshActiveCount() {
        const activeBtn = navContainer.querySelector('.recipe-tab-btn[data-id="active"]');
        if (activeBtn) {
            const activeCat = categories.find(c => c.id === 'active');
            updateButtonText(activeBtn, activeCat);
        }
    }

    // 초기 렌더링
    renderGrid(categories[0].items);
    refreshActiveCount();
}