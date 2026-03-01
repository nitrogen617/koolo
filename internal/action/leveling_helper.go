package action

import (
	"slices"
	"strings"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/lxn/win"
)

func ResetStaminaCooldown(ctx *context.Context) {
	if ctx == nil {
		return
	}
	ctx.LastStaminaPotUse = time.Time{}
	ctx.StaminaPotCooldown = 0
}

func TryBuyAndConsumeStaminaPots() {
	ctx := context.Get()

	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); !isLevelingChar {
		return
	}
	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return
	}
	lvl, found := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
	if !found || lvl.Value >= 18 {
		return
	}
	gold := ctx.Data.PlayerUnit.TotalPlayerGold()
	var targetCount int
	switch {
	case gold >= 2000:
		targetCount = 10
	case gold >= 1000:
		targetCount = 5
	default:
		return
	}

	existingCount := 0
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if itm.Name == "StaminaPotion" {
			existingCount++
		}
	}

	buyCount := targetCount - existingCount
	if buyCount < 0 {
		buyCount = 0
	}
	if !ctx.LastStaminaPotUse.IsZero() && time.Since(ctx.LastStaminaPotUse) < ctx.StaminaPotCooldown {
		return
	}

	if buyCount > 0 {
		vendorNPC := town.GetTownByArea(ctx.Data.PlayerUnit.Area).RefillNPC()
		if ctx.Data.PlayerUnit.Area.Act() == 2 {
			vendorNPC = npc.Lysander
		}
		if err := InteractNPC(vendorNPC); err != nil {
			return
		}
		defer step.CloseAllMenus()

		if vendorNPC == npc.Jamella {
			ctx.HID.KeySequence(win.VK_HOME, win.VK_RETURN)
		} else {
			ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
		}

		SwitchVendorTab(4)
		ctx.RefreshGameData()
		if staminaPot, found := ctx.Data.Inventory.Find(item.Name("StaminaPotion"), item.LocationVendor); found {
			town.BuyItem(staminaPot, buyCount)
		}
		step.CloseAllMenus()
	}

	if !ctx.Data.OpenMenus.Inventory {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		utils.Sleep(200)
	}

	ctx.RefreshInventory()
	availableCount := 0
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if itm.Name == "StaminaPotion" {
			availableCount++
		}
	}

	consumeCount := targetCount
	if availableCount < consumeCount {
		consumeCount = availableCount
	}
	if consumeCount == 0 {
		step.CloseAllMenus()
		return
	}

	for i := 0; i < consumeCount; i++ {
		ctx.RefreshInventory()
		pot, ok := ctx.Data.Inventory.Find(item.Name("StaminaPotion"), item.LocationInventory)
		if !ok {
			break
		}
		screenPos := ui.GetScreenCoordsForItem(pot)
		ctx.HID.Click(game.RightButton, screenPos.X, screenPos.Y)
		utils.Sleep(150)
	}
	step.CloseAllMenus()
	if consumeCount > 0 {
		ctx.LastStaminaPotUse = time.Now()
		ctx.StaminaPotCooldown = time.Duration(consumeCount) * 30 * time.Second
	}
}

// TryBuyLevelingBelt attempts to buy a better belt for leveling characters based on act and level.
func TryBuyLevelingBelt() {
	ctx := context.Get()
	if ctx == nil {
		return
	}
	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); !isLevelingChar {
		return
	}
	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return
	}

	switch ctx.Data.PlayerUnit.Area.Act() {
	case 1:
		_ = buyAct1LowLevelBeltFromCharsi(ctx)
		_ = gambleAct1Belt(ctx)
	case 2:
		_ = buyAct2Belt(ctx)
	case 4:
		_ = buyAct4PlatedBeltFromJamella(ctx)
	}
}

func gambleAct1Belt(ctx *context.Status) error {
	// Check if level 9. Some wiggle room for over leveling, but then stops for level 11+
	lvl, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value < 9 || lvl.Value >= 11 {
		ctx.Logger.Info("Not level 9 to 11, skipping belt gamble.")
		return nil
	}
	str, found := ctx.Data.PlayerUnit.FindStat(stat.Strength, 0)
	if !found || str.Value < 25 {
		ctx.Logger.Info("Not enough strength for a belt yet, skipping belt gamble.")
		return nil
	}

	// Check equipped and inventory for a suitable belt first
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		if itm.Name == "Belt" || itm.Name == "HeavyBelt" || itm.Name == "PlatedBelt" {
			ctx.Logger.Info("Already have a 9 slot belt equipped, skipping.")
			return nil
		}
	}
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if itm.Name == "Belt" || itm.Name == "HeavyBelt" || itm.Name == "PlatedBelt" {
			ctx.Logger.Info("Already have a 9 slot belt in inventory, skipping.")
			return nil
		}
	}

	// Check for gold before visiting the vendor
	if ctx.Data.PlayerUnit.TotalPlayerGold() < 3000 {
		ctx.Logger.Info("Not enough gold to buy a belt, skipping.")
		return nil
	}

	// Go to Gheed and get the gambling menu
	ctx.Logger.Info("No 12 slot belt found, trying to buy one from Gheed.")
	if err := InteractNPC(npc.Gheed); err != nil {
		return err
	}
	defer step.CloseAllMenus()

	ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_DOWN, win.VK_RETURN)
	utils.Sleep(1000)

	// Check if the shop menu is open
	if !ctx.Data.OpenMenus.NPCShop {
		ctx.Logger.Debug("failed opening gambling window")
		return nil
	}

	// Define the item to gamble for
	itemsToGamble := []string{"Belt"}

	const maxGambleRefreshes = 25
	// Loop until the desired item is found and purchased
	for attempt := 0; attempt < maxGambleRefreshes; attempt++ {
		// Check for any of the desired items in the vendor's inventory
		for _, itmName := range itemsToGamble {
			itm, found := ctx.Data.Inventory.Find(item.Name(itmName), item.LocationVendor)
			if found {
				town.BuyItem(itm, 1)
				ctx.RefreshGameData()
				ctx.Logger.Info("Belt purchased, equipping.")
				if belt, ok := ctx.Data.Inventory.Find(item.Name("Belt"), item.LocationInventory); ok {
					if err := EquipItem(belt, item.LocBelt, item.LocationEquipped); err != nil {
						ctx.Logger.Error("Failed to equip belt after buying", "error", err)
					}
					return nil
				}
				if belt, ok := ctx.Data.Inventory.Find(item.Name("HeavyBelt"), item.LocationInventory); ok {
					if err := EquipItem(belt, item.LocBelt, item.LocationEquipped); err != nil {
						ctx.Logger.Error("Failed to equip belt after buying", "error", err)
					}
					return nil
				}
				if belt, ok := ctx.Data.Inventory.Find(item.Name("PlatedBelt"), item.LocationInventory); ok {
					if err := EquipItem(belt, item.LocBelt, item.LocationEquipped); err != nil {
						ctx.Logger.Error("Failed to equip belt after buying", "error", err)
					}
					return nil
				}
				for _, belt := range ctx.Data.Inventory.ByLocation(item.LocationCursor) {
					if belt.Name != "Belt" && belt.Name != "HeavyBelt" && belt.Name != "PlatedBelt" {
						continue
					}
					if slotCoords, ok := getEquippedSlotCoords(item.LocBelt, ctx.Data.LegacyGraphics); ok {
						ctx.HID.Click(game.LeftButton, slotCoords.X, slotCoords.Y)
						utils.Sleep(300)
						ctx.RefreshGameData()
					}
					return nil
				}
				ctx.Logger.Warn("Purchased belt not found in inventory or cursor after buying")
				return nil
			}
		}

		// If no desired item was found, refresh the gambling window
		ctx.Logger.Info("Desired items not found in gambling window, refreshing...")
		if ctx.Data.LegacyGraphics {
			ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonXClassic, ui.GambleRefreshButtonYClassic)
		} else {
			ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonX, ui.GambleRefreshButtonY)
		}
		utils.Sleep(500)
	}
	ctx.Logger.Info("Belt not found at Gheed after gamble refreshes", "attempts", maxGambleRefreshes)
	return nil
}

func buyAct1LowLevelBeltFromCharsi(ctx *context.Status) error {
	lvl, found := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
	if !found || lvl.Value < 6 {
		return nil
	}
	if ctx.Data.PlayerUnit.TotalPlayerGold() < 200 || hasAnyBelt(ctx) {
		return nil
	}

	if err := InteractNPC(npc.Charsi); err != nil {
		return err
	}
	defer step.CloseAllMenus()

	ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
	SwitchVendorTab(1)
	purchasedBeltName := item.Name("")
	if beltItem, found := ctx.Data.Inventory.Find(item.Name("LightBelt"), item.LocationVendor); found {
		town.BuyItem(beltItem, 1)
		purchasedBeltName = item.Name("LightBelt")
	} else if beltItem, found := ctx.Data.Inventory.Find(item.Name("Sash"), item.LocationVendor); found {
		town.BuyItem(beltItem, 1)
		purchasedBeltName = item.Name("Sash")
	}
	ctx.RefreshGameData()
	if purchasedBeltName == "" {
		return nil
	}

	if belt, ok := ctx.Data.Inventory.Find(purchasedBeltName, item.LocationInventory); ok {
		if err := EquipItem(belt, item.LocBelt, item.LocationEquipped); err != nil {
			ctx.Logger.Error("Failed to equip belt after buying from Charsi", "error", err)
		}
		return nil
	}

	for _, belt := range ctx.Data.Inventory.ByLocation(item.LocationCursor) {
		if belt.Name != purchasedBeltName {
			continue
		}
		if slotCoords, ok := getEquippedSlotCoords(item.LocBelt, ctx.Data.LegacyGraphics); ok {
			ctx.HID.Click(game.LeftButton, slotCoords.X, slotCoords.Y)
			utils.Sleep(300)
			ctx.RefreshGameData()
		}
		return nil
	}

	ctx.Logger.Warn("Purchased Act 1 belt not found in inventory or cursor", "item", purchasedBeltName)

	return nil
}

func buyAct2Belt(ctx *context.Status) error {
	// Only buy belts in Normal difficulty
	if ctx.CharacterCfg.Game.Difficulty != difficulty.Normal {
		return nil
	}

	// Check equipped and inventory for a suitable belt first
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
		if itm.Name == "Belt" || itm.Name == "HeavyBelt" || itm.Name == "PlatedBelt" {
			ctx.Logger.Info("Already have a 12+ slot belt equipped, skipping.")
			return nil
		}
	}
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if itm.Name == "Belt" || itm.Name == "HeavyBelt" || itm.Name == "PlatedBelt" {
			ctx.Logger.Info("Already have a 12+ slot belt in inventory, skipping.")
			return nil
		}
	}

	// Check for gold before visiting the vendor
	if ctx.Data.PlayerUnit.TotalPlayerGold() < 1000 {
		ctx.Logger.Info("Not enough gold to buy a belt, skipping.")
		return nil
	}

	ctx.Logger.Info("No 12 slot belt found, trying to buy one from Fara.")
	if err := InteractNPC(npc.Fara); err != nil {
		return err
	}
	defer step.CloseAllMenus()

	ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN) // Interact with Fara
	utils.Sleep(1000)

	// Switch to armor tab and refresh game data to see the new items
	SwitchVendorTab(1)
	ctx.RefreshGameData()
	utils.Sleep(500)

	// Find a suitable belt to buy from vendor
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationVendor) {
		// We are looking for a "Belt", which has 12 slots.
		if itm.Name == "Belt" {
			strReq := itm.Desc().RequiredStrength
			ctx.Logger.Debug("Vendor item found", "name", itm.Name, "strReq", strReq)

			if strReq <= 25 {
				ctx.Logger.Info("Found a suitable belt, buying it.", "name", itm.Name)
				town.BuyItem(itm, 1)
				ctx.RefreshGameData()
				ctx.Logger.Info("Belt purchased, running AutoEquip.")
				if err := AutoEquip(); err != nil {
					ctx.Logger.Error("AutoEquip failed after buying belt", "error", err)
				}

				return nil
			}
		}
	}

	ctx.Logger.Info("No suitable belt found at Fara.")
	return nil
}

func buyAct4PlatedBeltFromJamella(ctx *context.Status) error {
	if !ctx.Data.IsLevelingCharacter {
		return nil
	}
	if !ctx.Data.PlayerUnit.Area.IsTown() || ctx.Data.PlayerUnit.Area.Act() != 4 {
		return nil
	}
	if ctx.Data.PlayerUnit.TotalPlayerGold() < 8000 {
		return nil
	}
	lvl, found := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
	if !found || lvl.Value > 30 {
		return nil
	}
	str, found := ctx.Data.PlayerUnit.FindStat(stat.Strength, 0)
	if !found || str.Value < 60 {
		return nil
	}
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationEquipped, item.LocationInventory) {
		if itm.Name == "PlatedBelt" {
			return nil
		}
	}

	if err := InteractNPC(npc.Jamella); err != nil {
		return err
	}
	defer step.CloseAllMenus()

	// Jamella gamble button is the second one.
	ctx.HID.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
	if !ctx.Data.OpenMenus.NPCShop {
		ctx.Logger.Debug("failed opening gambling window for Jamella")
		return nil
	}

	const maxGambleRefreshes = 25
	for attempt := 0; attempt < maxGambleRefreshes; attempt++ {
		ctx.RefreshGameData()
		if beltItem, found := ctx.Data.Inventory.Find(item.Name("PlatedBelt"), item.LocationVendor); found {
			town.BuyItem(beltItem, 1)
			ctx.RefreshGameData()
			if belt, ok := ctx.Data.Inventory.Find(item.Name("PlatedBelt"), item.LocationInventory); ok {
				if err := EquipItem(belt, item.LocBelt, item.LocationEquipped); err != nil {
					ctx.Logger.Error("Failed to equip plated belt after gambling", "error", err)
				}
				return nil
			}
			for _, belt := range ctx.Data.Inventory.ByLocation(item.LocationCursor) {
				if belt.Name != "PlatedBelt" {
					continue
				}
				if slotCoords, ok := getEquippedSlotCoords(item.LocBelt, ctx.Data.LegacyGraphics); ok {
					ctx.HID.Click(game.LeftButton, slotCoords.X, slotCoords.Y)
					utils.Sleep(300)
					ctx.RefreshGameData()
				}
				return nil
			}
			ctx.Logger.Warn("PlatedBelt purchased but not found in inventory or cursor")
			return nil
		}
		if ctx.Data.LegacyGraphics {
			ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonXClassic, ui.GambleRefreshButtonYClassic)
		} else {
			ctx.HID.Click(game.LeftButton, ui.GambleRefreshButtonX, ui.GambleRefreshButtonY)
		}
		utils.Sleep(500)
	}
	ctx.Logger.Info("PlatedBelt not found at Jamella after gamble refreshes", "attempts", maxGambleRefreshes)

	return nil
}

func hasAnyBelt(ctx *context.Status) bool {
	for _, itm := range ctx.Data.Inventory.ByLocation(item.LocationEquipped, item.LocationInventory) {
		if itm.Name == "PlatedBelt" || itm.Name == "HeavyBelt" || itm.Name == "Belt" || itm.Name == "LightBelt" || itm.Name == "Sash" {
			return true
		}
	}
	return false
}

const levelingSocketGemKeepLimit = 3
const levelingHelmetRalReserve = 1
const levelingHelmetRalMinLevel = 19

func levelingGemSearchLocations(ctx *context.Status) []item.LocationType {
	locations := []item.LocationType{
		item.LocationInventory,
		item.LocationStash,
		item.LocationSharedStash,
	}
	if ctx != nil && ctx.Data.IsDLC() {
		locations = append(locations, item.LocationGemsTab, item.LocationRunesTab)
	}

	return locations
}

func availableLevelingGemQuantity(itm data.Item) int {
	if itm.Location.LocationType == item.LocationGemsTab || itm.Location.LocationType == item.LocationRunesTab {
		qty := isDLCStackedQuantity(itm)
		if qty > 0 {
			return qty
		}
	}

	return 1
}

func levelingSocketGemKind(itm data.Item) (string, int, bool) {
	for _, kind := range []string{"Diamond", "Ruby", "Sapphire"} {
		if rank, ok := gemRank(string(itm.Name), kind); ok {
			return kind, rank, true
		}
	}

	return "", 0, false
}

func isLevelingSocketGem(itm data.Item) bool {
	_, _, ok := levelingSocketGemKind(itm)
	return ok
}

func levelingSocketGemStorageLocations(ctx *context.Status) []item.LocationType {
	locations := []item.LocationType{
		item.LocationStash,
		item.LocationSharedStash,
	}
	if ctx != nil && ctx.Data.IsDLC() {
		locations = append(locations, item.LocationGemsTab)
	}

	return locations
}

func RebalanceLevelingSocketGemStash() {
	ctx := context.Get()
	openedStash := false

	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); !isLevelingChar {
		return
	}
	if !ctx.Data.PlayerUnit.Area.IsTown() {
		return
	}
	if ctx.Data.IsDLC() {
		return
	}
	if !ctx.Data.OpenMenus.Stash {
		if err := OpenStash(); err != nil {
			ctx.Logger.Debug("Leveling gem rebalance skipped: failed opening stash", "error", err)
			return
		}
		openedStash = true
		defer step.CloseAllMenus()
	}

	ctx.RefreshGameData()
	ctx.RefreshInventory()

	type storedGem struct {
		item data.Item
		kind string
		rank int
		qty  int
	}

	gemsByKind := map[string][]storedGem{}
	for _, itm := range FilterDLCGhostItems(ctx.Data.Inventory.ByLocation(levelingSocketGemStorageLocations(ctx)...)) {
		kind, rank, ok := levelingSocketGemKind(itm)
		if !ok {
			continue
		}
		qty := availableLevelingGemQuantity(itm)
		if qty <= 0 {
			continue
		}
		gemsByKind[kind] = append(gemsByKind[kind], storedGem{
			item: itm,
			kind: kind,
			rank: rank,
			qty:  qty,
		})
	}

	for kind, stored := range gemsByKind {
		totalQty := 0
		for _, gem := range stored {
			totalQty += gem.qty
		}
		if totalQty <= levelingSocketGemKeepLimit {
			continue
		}

		excess := totalQty - levelingSocketGemKeepLimit
		slices.SortFunc(stored, func(a, b storedGem) int {
			if a.rank != b.rank {
				return a.rank - b.rank
			}
			return a.qty - b.qty
		})

		moved := 0
		for _, candidate := range stored {
			for candidate.qty > 0 && excess > 0 {
				currentGem, found := ctx.Data.Inventory.FindByID(candidate.item.UnitID)
				if !found {
					break
				}
				if !inventoryCanFitItem(currentGem) {
					ctx.Logger.Debug("Not enough inventory space to rebalance leveling gems", "kind", kind, "remainingExcess", excess)
					break
				}
				if _, ok := moveItemToInventory(currentGem); !ok {
					ctx.Logger.Debug("Failed to move excess leveling gem to inventory", "kind", kind, "gem", currentGem.Name)
					break
				}
				excess--
				candidate.qty--
				moved++
				ctx.RefreshInventory()
			}
			if excess <= 0 {
				break
			}
		}

		if moved > 0 {
			ctx.Logger.Debug("Rebalanced leveling gems", "kind", kind, "movedToInventory", moved, "kept", totalQty-moved)
		}
	}

	if openedStash {
		ctx.RefreshGameData()
	}
}

func TrySocketLevelingGems() {
	ctx := context.Get()

	if _, isLevelingChar := ctx.Char.(context.LevelingCharacter); !isLevelingChar {
		ctx.Logger.Debug("Leveling socket skipped: character is not a leveling build")
		return
	}
	if !ctx.Data.PlayerUnit.Area.IsTown() {
		ctx.Logger.Debug("Leveling socket skipped: player is not in town", "area", ctx.Data.PlayerUnit.Area)
		return
	}
	ctx.Logger.Debug("Leveling socket evaluation started")
	gemSearchLocations := levelingGemSearchLocations(ctx)

	hasSocketCandidate := false
	hasStashCandidate := false
	hasStashGems := false
	isSocketCandidate := func(candidate data.Item) bool {
		if candidate.IsRuneword || candidate.IsBroken {
			return false
		}
		numSockets, found := candidate.FindStat(stat.NumSockets, 0)
		if !found || numSockets.Value <= len(candidate.Sockets) {
			return false
		}
		isShield := candidate.Type().IsType(item.TypeShield)
		isArmor := candidate.Type().IsType(item.TypeArmor)
		isHelm := candidate.Type().IsType(item.TypeHelm)
		return isShield || isArmor || isHelm
	}
	for _, candidate := range ctx.Data.Inventory.ByLocation(item.LocationEquipped, item.LocationInventory) {
		if isSocketCandidate(candidate) {
			hasSocketCandidate = true
			break
		}
	}
	for _, candidate := range ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash) {
		if isSocketCandidate(candidate) {
			hasStashCandidate = true
			break
		}
	}
	if !hasSocketCandidate && !hasStashCandidate {
		ctx.Logger.Debug("Leveling socket skipped: no socketable base found")
		return
	}
	for _, gem := range ctx.Data.Inventory.ByLocation(gemSearchLocations...) {
		if _, ok := gemRank(string(gem.Name), "Ruby"); ok {
			hasStashGems = true
			break
		}
		if _, ok := gemRank(string(gem.Name), "Sapphire"); ok {
			hasStashGems = true
			break
		}
		if _, ok := gemRank(string(gem.Name), "Diamond"); ok {
			hasStashGems = true
			break
		}
		if gem.Name == "RalRune" {
			hasStashGems = true
			break
		}
	}

	includeStash := hasStashCandidate || hasStashGems
	if includeStash {
		ctx.RefreshGameData()
		ctx.RefreshInventory()
	}
	gemPool := ctx.Data.Inventory.ByLocation(gemSearchLocations...)
	if len(gemPool) == 0 {
		ctx.Logger.Debug("Leveling socket skipped: no gems or runes found in search locations")
		return
	}
	usedGemCount := map[data.UnitID]int{}
	playerLevel := 0
	if lvl, found := ctx.Data.PlayerUnit.FindStat(stat.Level, 0); found {
		playerLevel = lvl.Value
	}

	gemRanks := func(kind string) []int {
		ranks := make([]int, 0, len(gemPool))
		for _, itm := range gemPool {
			qty := availableLevelingGemQuantity(itm) - usedGemCount[itm.UnitID]
			if qty <= 0 {
				continue
			}
			rank, ok := gemRank(string(itm.Name), kind)
			if ok {
				for i := 0; i < qty; i++ {
					ranks = append(ranks, rank)
				}
			}
		}
		return ranks
	}

	findBestGem := func(kind string) (data.Item, bool) {
		bestRank := 0
		bestItem := data.Item{}
		bestLocationPriority := 99
		for _, itm := range gemPool {
			qty := availableLevelingGemQuantity(itm) - usedGemCount[itm.UnitID]
			if qty <= 0 {
				continue
			}
			if kind == "RalRune" {
				if itm.Name != item.Name(kind) {
					continue
				}
				remainingRal := 0
				for _, candidate := range gemPool {
					qty := availableLevelingGemQuantity(candidate) - usedGemCount[candidate.UnitID]
					if qty <= 0 || candidate.Name != item.Name("RalRune") {
						continue
					}
					remainingRal += qty
				}
				if remainingRal <= levelingHelmetRalReserve {
					return data.Item{}, false
				}
				locationPriority := 2
				switch itm.Location.LocationType {
				case item.LocationInventory:
					locationPriority = 0
				case item.LocationStash, item.LocationSharedStash:
					locationPriority = 1
				}
				if bestItem.UnitID == 0 || locationPriority < bestLocationPriority {
					bestItem = itm
					bestLocationPriority = locationPriority
				}
				continue
			}
			rank, ok := gemRank(string(itm.Name), kind)
			if !ok {
				continue
			}
			if rank > bestRank || (rank == bestRank && itm.Location.LocationType == item.LocationInventory && bestItem.Location.LocationType != item.LocationInventory) {
				bestRank = rank
				bestItem = itm
			}
		}
		if kind == "RalRune" {
			if bestItem.UnitID == 0 {
				return data.Item{}, false
			}
			return bestItem, true
		}
		if bestRank == 0 {
			return data.Item{}, false
		}
		return bestItem, true
	}

	buildGemPlan := func(isShield, isArmor, isHelm bool, sockets int) []string {
		if sockets <= 0 {
			return nil
		}
		plan := make([]string, 0, sockets)
		if isShield {
			diamondCount := len(gemRanks("Diamond"))
			for i := 0; i < sockets && diamondCount > 0; i++ {
				plan = append(plan, "Diamond")
				diamondCount--
			}
			return plan
		}
		if isHelm && playerLevel >= levelingHelmetRalMinLevel {
			availableRals := 0
			for _, itm := range gemPool {
				qty := availableLevelingGemQuantity(itm) - usedGemCount[itm.UnitID]
				if qty <= 0 || itm.Name != "RalRune" {
					continue
				}
				availableRals += qty
			}
			usableRals := availableRals - levelingHelmetRalReserve
			if usableRals < 0 {
				usableRals = 0
			}
			for i := 0; i < sockets && usableRals > 0; i++ {
				plan = append(plan, "RalRune")
				usableRals--
			}
		}
		if isArmor || isHelm {
			rubyCount := len(gemRanks("Ruby"))
			sapphireCount := len(gemRanks("Sapphire"))
			for len(plan) < sockets && rubyCount > 0 {
				plan = append(plan, "Ruby")
				rubyCount--
			}
			for len(plan) < sockets && sapphireCount > 0 {
				plan = append(plan, "Sapphire")
				sapphireCount--
			}
		}
		return plan
	}
	buildHelmFallbackPlan := func(sockets int) []string {
		if sockets <= 0 {
			return nil
		}
		plan := make([]string, 0, sockets)
		rubyCount := len(gemRanks("Ruby"))
		sapphireCount := len(gemRanks("Sapphire"))
		for len(plan) < sockets && rubyCount > 0 {
			plan = append(plan, "Ruby")
			rubyCount--
		}
		for len(plan) < sockets && sapphireCount > 0 {
			plan = append(plan, "Sapphire")
			sapphireCount--
		}
		return plan
	}

	changed := false
	const minGemUpgradeScore = 5.0
	candidates := ctx.Data.Inventory.ByLocation(item.LocationEquipped, item.LocationInventory)
	if includeStash {
		candidates = append(candidates, ctx.Data.Inventory.ByLocation(item.LocationStash, item.LocationSharedStash)...)
	}

	for _, candidate := range candidates {
		if candidate.IsRuneword || candidate.IsBroken {
			continue
		}
		numSockets, found := candidate.FindStat(stat.NumSockets, 0)
		if !found || numSockets.Value <= len(candidate.Sockets) {
			continue
		}

		isShield := candidate.Type().IsType(item.TypeShield)
		isArmor := candidate.Type().IsType(item.TypeArmor)
		isHelm := candidate.Type().IsType(item.TypeHelm)
		if !isShield && !isArmor && !isHelm {
			continue
		}

		openSockets := numSockets.Value - len(candidate.Sockets)
		targetLoc := candidate.Location.BodyLocation
		if candidate.Location.LocationType != item.LocationEquipped {
			switch {
			case isShield:
				targetLoc = item.LocRightArm
			case isArmor:
				targetLoc = item.LocTorso
			case isHelm:
				targetLoc = item.LocHead
			}
		}
		gemPlan := buildGemPlan(isShield, isArmor, isHelm, openSockets)
		if len(gemPlan) == 0 {
			ctx.Logger.Debug("Leveling socket skipped candidate: no plan built",
				"item", candidate.Name,
				"unitID", candidate.UnitID,
				"openSockets", openSockets,
			)
			continue
		}
		predictedScore := predictedScoreWithAvailableSocketables(candidate, isShield, isArmor, isHelm, openSockets, gemRanks, gemPlan)
		if !isPredictedGemUpgrade(candidate, predictedScore, minGemUpgradeScore) {
			if isHelm && slices.Contains(gemPlan, "RalRune") {
				fallbackPlan := buildHelmFallbackPlan(openSockets)
				if len(fallbackPlan) == 0 {
					ctx.Logger.Debug("Leveling socket helm fallback unavailable after Ral plan scored too low",
						"item", candidate.Name,
						"unitID", candidate.UnitID,
						"originalPlan", strings.Join(gemPlan, ","),
					)
				}
				if len(fallbackPlan) > 0 {
					fallbackPredicted := predictedScoreWithAvailableSocketables(candidate, isShield, isArmor, isHelm, openSockets, gemRanks, fallbackPlan)
					if isPredictedGemUpgrade(candidate, fallbackPredicted, minGemUpgradeScore) {
						ctx.Logger.Debug("Leveling socket switching helm plan after Ral candidate scored too low",
							"item", candidate.Name,
							"unitID", candidate.UnitID,
							"originalPlan", strings.Join(gemPlan, ","),
							"fallbackPlan", strings.Join(fallbackPlan, ","),
						)
						gemPlan = fallbackPlan
						predictedScore = fallbackPredicted
					} else {
						ctx.Logger.Debug("Leveling socket helm fallback also scored too low",
							"item", candidate.Name,
							"unitID", candidate.UnitID,
							"originalPlan", strings.Join(gemPlan, ","),
							"fallbackPlan", strings.Join(fallbackPlan, ","),
						)
					}
				}
			}
		}
		if !isPredictedGemUpgrade(candidate, predictedScore, minGemUpgradeScore) {
			ctx.Logger.Debug("Leveling socket skipped candidate: predicted upgrade too small",
				"item", candidate.Name,
				"unitID", candidate.UnitID,
				"openSockets", openSockets,
				"plan", strings.Join(gemPlan, ","),
			)
			continue
		}
		ctx.Logger.Debug("Leveling socket candidate selected",
			"item", candidate.Name,
			"unitID", candidate.UnitID,
			"location", candidate.Location.LocationType,
			"openSockets", openSockets,
			"plan", strings.Join(gemPlan, ","),
		)
		wasEquipped := candidate.Location.LocationType == item.LocationEquipped
		updatedCandidate := candidate
		socketed := false
		attemptedSell := false
		movedFromEquip := false
		gemPlanIdx := 0
		for openSockets > 0 {
			if gemPlanIdx >= len(gemPlan) {
				break
			}
			if !ensureInventorySpaceForSocketing(updatedCandidate) {
				if attemptedSell {
					ctx.Logger.Debug("Leveling socket aborted after retry because inventory space is still unavailable",
						"item", updatedCandidate.Name,
						"unitID", updatedCandidate.UnitID,
					)
					break
				}
				attemptedSell = true
				VendorRefill(VendorRefillOpts{SellJunk: true})
				if !ctx.Data.OpenMenus.Inventory {
					ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
					utils.Sleep(200)
				}
				ctx.RefreshInventory()
				if !ensureInventorySpaceForSocketing(updatedCandidate) {
					ctx.Logger.Debug("Leveling socket inventory space check failed after vendor sell",
						"item", updatedCandidate.Name,
						"unitID", updatedCandidate.UnitID,
					)
					break
				}
			}
			kind := gemPlan[gemPlanIdx]
			gemPlanIdx++
			gem, ok := findBestGem(kind)
			if !ok {
				ctx.Logger.Debug("Leveling socket could not find planned socketable",
					"item", updatedCandidate.Name,
					"unitID", updatedCandidate.UnitID,
					"planned", kind,
				)
				break
			}
			ctx.Logger.Debug("Leveling socket attempting insert",
				"item", updatedCandidate.Name,
				"unitID", updatedCandidate.UnitID,
				"socketable", gem.Name,
				"socketableUnitID", gem.UnitID,
				"socketableLocation", gem.Location.LocationType,
			)

			updatedItem := updatedCandidate
			if updatedCandidate.Location.LocationType != item.LocationEquipped {
				updatedItem, ok = moveItemToInventory(updatedCandidate)
				if !ok {
					ctx.Logger.Debug("Leveling socket failed moving base item to inventory",
						"item", updatedCandidate.Name,
						"unitID", updatedCandidate.UnitID,
						"location", updatedCandidate.Location.LocationType,
					)
					break
				}
				movedFromEquip = wasEquipped
			}

			if !insertGemIntoItem(gem, updatedItem) {
				ctx.Logger.Debug("Leveling socket failed inserting socketable into item",
					"item", updatedItem.Name,
					"unitID", updatedItem.UnitID,
					"socketable", gem.Name,
					"socketableUnitID", gem.UnitID,
				)
				if movedFromEquip && targetLoc != item.LocationUnknown {
					_ = EquipItem(updatedItem, targetLoc, item.LocationEquipped)
					utils.Sleep(300)
				}
				break
			}

			usedGemCount[gem.UnitID]++
			ctx.RefreshInventory()
			updatedAfter, found := ctx.Data.Inventory.FindByID(updatedItem.UnitID)
			if !found {
				ctx.Logger.Debug("Leveling socket lost track of base item after insert",
					"item", updatedItem.Name,
					"unitID", updatedItem.UnitID,
				)
				if movedFromEquip && targetLoc != item.LocationUnknown {
					_ = EquipItem(updatedItem, targetLoc, item.LocationEquipped)
					utils.Sleep(300)
				}
				break
			}

			updatedCandidate = updatedAfter
			socketed = true
			changed = true
			ctx.Logger.Debug("Leveling socket inserted successfully",
				"item", updatedAfter.Name,
				"unitID", updatedAfter.UnitID,
				"used", gem.Name,
			)

			if sockets, found := updatedAfter.FindStat(stat.NumSockets, 0); found {
				newOpen := sockets.Value - len(updatedAfter.Sockets)
				if newOpen < 0 {
					newOpen = 0
				}
				if newOpen >= openSockets {
					openSockets--
				} else {
					openSockets = newOpen
				}
			} else {
				openSockets--
			}
		}

		if socketed {
			shouldEquip := wasEquipped
			if !shouldEquip {
				if targetLoc != item.LocationUnknown {
					equipped := GetEquippedItem(ctx.Data.Inventory, targetLoc)
					updatedScore, ok := PlayerScore(updatedCandidate)[targetLoc]
					if ok && equipped.UnitID == 0 {
						shouldEquip = true
					} else if ok {
						if equippedScore, ok := PlayerScore(equipped)[targetLoc]; ok && updatedScore > equippedScore {
							shouldEquip = true
						}
					}
				}
			}

			if shouldEquip && targetLoc != item.LocationUnknown && updatedCandidate.Location.LocationType != item.LocationEquipped {
				_ = EquipItem(updatedCandidate, targetLoc, item.LocationEquipped)
				utils.Sleep(300)
			}
		} else if movedFromEquip && updatedCandidate.Location.LocationType != item.LocationEquipped && targetLoc != item.LocationUnknown {
			_ = EquipItem(updatedCandidate, targetLoc, item.LocationEquipped)
			utils.Sleep(300)
		}
	}

	if err := step.CloseAllMenus(); err != nil {
		ctx.Logger.Debug("Leveling socket final menu cleanup failed", "error", err)
	}
	if changed {
		_ = AutoEquip()
	}
}

func gemRank(name string, kind string) (int, bool) {
	if !strings.HasSuffix(name, kind) {
		return 0, false
	}
	switch {
	case strings.HasPrefix(name, "Perfect"):
		return 5, true
	case strings.HasPrefix(name, "Flawless"):
		return 4, true
	case name == kind:
		return 3, true
	case strings.HasPrefix(name, "Flawed"):
		return 2, true
	case strings.HasPrefix(name, "Chipped"):
		return 1, true
	default:
		return 0, false
	}
}

func isPredictedGemUpgrade(candidate data.Item, predicted map[item.LocationType]float64, minUpgradeScore float64) bool {
	if len(predicted) == 0 {
		return false
	}

	ctx := context.Get()
	if candidate.Location.LocationType == item.LocationEquipped {
		current := PlayerScore(candidate)
		loc := candidate.Location.BodyLocation
		predictedScore, ok := predicted[loc]
		if !ok {
			return false
		}
		return predictedScore-current[loc] >= minUpgradeScore
	}

	for loc, predictedScore := range predicted {
		equipped := GetEquippedItem(ctx.Data.Inventory, loc)
		if equipped.UnitID == 0 {
			if predictedScore >= minUpgradeScore {
				return true
			}
			continue
		}
		equippedScore := PlayerScore(equipped)
		if predictedScore-equippedScore[loc] >= minUpgradeScore {
			return true
		}
	}

	return false
}

func predictedScoreWithAvailableSocketables(candidate data.Item, isShield, isArmor, isHelm bool, openSockets int, gemRanks func(kind string) []int, socketPlan []string) map[item.LocationType]float64 {
	if openSockets <= 0 {
		return PlayerScore(candidate)
	}

	lifeByRank := []int{0, 10, 17, 24, 31, 38}
	manaByRank := []int{0, 10, 17, 24, 31, 38}
	resByRank := []int{0, 6, 8, 11, 14, 19}
	bonusByStat := map[stat.ID]int{}

	if isShield {
		diamondRanks := gemRanks("Diamond")
		applyBestGemBonuses(bonusByStat, diamondRanks, openSockets, resByRank, stat.FireResist, stat.ColdResist, stat.LightningResist, stat.PoisonResist)
	} else if isArmor || isHelm {
		if isHelm && len(socketPlan) > 0 {
			// Ral provides +30 fire resist in helms. The plan already reserves one spare rune.
			plannedRals := 0
			for _, planned := range socketPlan {
				if planned == "RalRune" {
					plannedRals++
				}
			}
			bonusByStat[stat.FireResist] += plannedRals * 30
			openSockets -= plannedRals
		}
		rubyRanks := gemRanks("Ruby")
		if len(rubyRanks) > 0 {
			used := applyBestGemBonuses(bonusByStat, rubyRanks, openSockets, lifeByRank, stat.MaxLife)
			openSockets -= used
		}
		if openSockets > 0 {
			sapphireRanks := gemRanks("Sapphire")
			applyBestGemBonuses(bonusByStat, sapphireRanks, openSockets, manaByRank, stat.MaxMana)
		}
	}

	if len(bonusByStat) == 0 {
		return PlayerScore(candidate)
	}

	merged := make(stat.Stats, 0, len(candidate.Stats)+len(bonusByStat))
	merged = append(merged, candidate.Stats...)
	for id, value := range bonusByStat {
		if value == 0 {
			continue
		}
		merged = append(merged, stat.Data{ID: id, Value: value})
	}

	tmp := candidate
	tmp.Stats = merged
	return PlayerScore(tmp)
}

func applyBestGemBonuses(bonusByStat map[stat.ID]int, ranks []int, sockets int, valuesByRank []int, statIDs ...stat.ID) int {
	if sockets <= 0 || len(ranks) == 0 {
		return 0
	}
	slices.Sort(ranks)
	used := 0
	for i := len(ranks) - 1; i >= 0 && used < sockets; i-- {
		rank := ranks[i]
		if rank <= 0 || rank >= len(valuesByRank) {
			continue
		}
		value := valuesByRank[rank]
		if value <= 0 {
			continue
		}
		for _, id := range statIDs {
			bonusByStat[id] += value
		}
		used++
	}
	return used
}

func inventoryCanFitItem(i data.Item) bool {
	invMatrix := context.Get().Data.Inventory.Matrix()

	for y := 0; y <= len(invMatrix)-i.Desc().InventoryHeight; y++ {
		for x := 0; x <= len(invMatrix[0])-i.Desc().InventoryWidth; x++ {
			freeSpace := true
			for dy := 0; dy < i.Desc().InventoryHeight; dy++ {
				for dx := 0; dx < i.Desc().InventoryWidth; dx++ {
					if invMatrix[y+dy][x+dx] {
						freeSpace = false
						break
					}
				}
				if !freeSpace {
					break
				}
			}

			if freeSpace {
				return true
			}
		}
	}

	return false
}

func ensureInventorySpaceForSocketing(itm data.Item) bool {
	if itm.Location.LocationType == item.LocationEquipped {
		return true
	}
	if itm.Location.LocationType == item.LocationStash || itm.Location.LocationType == item.LocationSharedStash || itm.Location.LocationType == item.LocationGemsTab || itm.Location.LocationType == item.LocationRunesTab {
		return inventoryCanFitItem(itm)
	}
	return true
}

func moveItemToInventory(itm data.Item) (data.Item, bool) {
	ctx := context.Get()
	if itm.Location.LocationType != item.LocationEquipped {
		if itm.Location.LocationType == item.LocationStash || itm.Location.LocationType == item.LocationSharedStash || itm.Location.LocationType == item.LocationGemsTab || itm.Location.LocationType == item.LocationRunesTab {
			if !ctx.Data.OpenMenus.Stash {
				if err := OpenStash(); err != nil {
					return data.Item{}, false
				}
			}
			if itm.Location.LocationType == item.LocationSharedStash {
				SwitchStashTab(itm.Location.Page + 1)
			} else if itm.Location.LocationType == item.LocationGemsTab {
				SwitchStashTab(StashTabGems)
			} else if itm.Location.LocationType == item.LocationRunesTab {
				SwitchStashTab(StashTabRunes)
			} else {
				SwitchStashTab(1)
			}
			utils.Sleep(200)
			knownInvItems := make(map[data.UnitID]struct{})
			for _, invItem := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
				knownInvItems[invItem.UnitID] = struct{}{}
			}
			screenPos := ui.GetScreenCoordsForItem(itm)
			ctx.HID.ClickWithModifier(game.LeftButton, screenPos.X, screenPos.Y, game.CtrlKey)
			utils.Sleep(300)
			ctx.RefreshInventory()
			if itm.Location.LocationType == item.LocationGemsTab || itm.Location.LocationType == item.LocationRunesTab {
				for _, invItem := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
					if _, known := knownInvItems[invItem.UnitID]; known {
						continue
					}
					if invItem.Name == itm.Name {
						return invItem, true
					}
				}
				return data.Item{}, false
			}
			updated, found := ctx.Data.Inventory.FindByID(itm.UnitID)
			if !found || updated.Location.LocationType != item.LocationInventory {
				return data.Item{}, false
			}
			return updated, true
		}
		return itm, true
	}
	if !ctx.Data.OpenMenus.Inventory {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		utils.Sleep(200)
		ctx.RefreshInventory()
	}
	slotCoords, found := getEquippedSlotCoords(itm.Location.BodyLocation, ctx.Data.LegacyGraphics)
	if !found {
		return data.Item{}, false
	}
	ctx.HID.ClickWithModifier(game.LeftButton, slotCoords.X, slotCoords.Y, game.CtrlKey)
	utils.Sleep(300)
	ctx.RefreshInventory()
	updated, found := ctx.Data.Inventory.FindByID(itm.UnitID)
	if !found || updated.Location.LocationType != item.LocationInventory {
		return data.Item{}, false
	}
	return updated, true
}

func insertGemIntoItem(gem data.Item, base data.Item) bool {
	ctx := context.Get()
	success := false
	defer func() {
		if success {
			return
		}
		if err := step.CloseAllMenus(); err != nil {
			ctx.Logger.Debug("Leveling socket cleanup failed after insert error", "error", err)
		}
		ctx.RefreshGameData()
	}()

	if gem.Location.LocationType != item.LocationInventory &&
		gem.Location.LocationType != item.LocationStash &&
		gem.Location.LocationType != item.LocationSharedStash &&
		gem.Location.LocationType != item.LocationGemsTab &&
		gem.Location.LocationType != item.LocationRunesTab {
		return false
	}
	if !ctx.Data.OpenMenus.Inventory {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
		utils.Sleep(200)
		ctx.RefreshInventory()
	}

	if gem.Location.LocationType == item.LocationGemsTab || gem.Location.LocationType == item.LocationRunesTab {
		if !ctx.Data.OpenMenus.Stash {
			if err := OpenStash(); err != nil {
				return false
			}
		}
		if gem.Location.LocationType == item.LocationRunesTab {
			SwitchStashTab(StashTabRunes)
		} else {
			SwitchStashTab(StashTabGems)
		}
		utils.Sleep(200)

		knownInvItems := make(map[data.UnitID]struct{})
		for _, invItem := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			knownInvItems[invItem.UnitID] = struct{}{}
		}

		gemPos := ui.GetScreenCoordsForItem(gem)
		ctx.HID.ClickWithModifier(game.LeftButton, gemPos.X, gemPos.Y, game.CtrlKey)
		utils.Sleep(300)
		ctx.RefreshInventory()

		movedGem := data.Item{}
		foundMovedGem := false
		for _, invItem := range ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			if _, known := knownInvItems[invItem.UnitID]; known {
				continue
			}
			if invItem.Name == gem.Name {
				movedGem = invItem
				foundMovedGem = true
				break
			}
		}
		if !foundMovedGem {
			return false
		}
		gem = movedGem
	}

	if gem.Location.LocationType == item.LocationStash || gem.Location.LocationType == item.LocationSharedStash {
		if !ctx.Data.OpenMenus.Stash {
			if err := OpenStash(); err != nil {
				return false
			}
		}
		if gem.Location.LocationType == item.LocationSharedStash {
			SwitchStashTab(gem.Location.Page + 1)
		} else {
			SwitchStashTab(1)
		}
		utils.Sleep(200)
	}

	gemPos := ui.GetScreenCoordsForItem(gem)
	pickedToCursor := false
	for attempt := 0; attempt < 3; attempt++ {
		ctx.HID.Click(game.LeftButton, gemPos.X, gemPos.Y)
		utils.Sleep(200)
		ctx.RefreshInventory()
		for _, cursorItem := range ctx.Data.Inventory.ByLocation(item.LocationCursor) {
			if cursorItem.UnitID == gem.UnitID || cursorItem.Name == gem.Name {
				pickedToCursor = true
				break
			}
		}
		if pickedToCursor {
			break
		}
	}
	if !pickedToCursor {
		return false
	}

	if base.Location.LocationType == item.LocationEquipped && ctx.Data.OpenMenus.Stash {
		_ = CloseStash()
		utils.Sleep(200)
		ctx.RefreshGameData()
		if ctx.Data.OpenMenus.Stash {
			return false
		}
		if !ctx.Data.OpenMenus.Inventory {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.Inventory)
			utils.Sleep(200)
		}
		ctx.RefreshInventory()
	}

	for attempt := 0; attempt < 2; attempt++ {
		switch base.Location.LocationType {
		case item.LocationInventory:
			basePos := ui.GetScreenCoordsForItem(base)
			ctx.HID.Click(game.LeftButton, basePos.X, basePos.Y)
		case item.LocationEquipped:
			slotCoords, found := getEquippedSlotCoords(base.Location.BodyLocation, ctx.Data.LegacyGraphics)
			if !found {
				return false
			}
			ctx.HID.Click(game.LeftButton, slotCoords.X, slotCoords.Y)
		default:
			return false
		}
		utils.Sleep(250)
		ctx.RefreshInventory()
		if len(ctx.Data.Inventory.ByLocation(item.LocationCursor)) == 0 {
			success = true
			return true
		}
	}

	DropAndRecoverCursorItem()
	return false
}
