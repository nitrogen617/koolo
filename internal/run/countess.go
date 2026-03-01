package run

import (
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

var nightmareCountessRunewordRequirements = map[item.RunewordName]map[string]int{
	item.RunewordStealth: {
		"TalRune": 1,
		"EthRune": 1,
	},
	item.RunewordSpirit: {
		"TalRune":  1,
		"ThulRune": 1,
		"OrtRune":  1,
		"AmnRune":  1,
	},
	item.RunewordInsight: {
		"RalRune": 1,
		"TalRune": 1,
		"SolRune": 1,
	},
	item.RunewordLore: {
		"OrtRune": 1,
		"SolRune": 1,
	},
}

var nightmareCountessTrackedRunes = map[string]struct{}{
	"TalRune":  {},
	"EthRune":  {},
	"ThulRune": {},
	"OrtRune":  {},
	"AmnRune":  {},
	"RalRune":  {},
	"SolRune":  {},
}

func selectedCountessRunewords(settings *SequenceSettings, currentDifficulty difficulty.Difficulty) []item.RunewordName {
	if settings == nil {
		return nil
	}

	if currentDifficulty == difficulty.Normal {
		if settings.SkipCountessWhenStealthReady {
			return []item.RunewordName{item.RunewordStealth}
		}
		return nil
	}

	if currentDifficulty != difficulty.Nightmare {
		return nil
	}

	selected := make([]item.RunewordName, 0, len(settings.CountessNightmareRunewords))
	for _, name := range settings.CountessNightmareRunewords {
		runeword := item.RunewordName(name)
		if _, isSupported := nightmareCountessRunewordRequirements[runeword]; isSupported && !slices.Contains(selected, runeword) {
			selected = append(selected, runeword)
		}
	}

	if len(selected) == 0 && settings.SkipCountessWhenNmCoreRunesReady {
		return []item.RunewordName{item.RunewordSpirit, item.RunewordInsight, item.RunewordLore}
	}

	return selected
}

func buildCountessInventory(items []data.Item) (map[string]int, map[item.RunewordName]bool) {
	runes := make(map[string]int)
	runewords := make(map[item.RunewordName]bool)

	for _, itm := range action.FilterDLCGhostItems(items) {
		if itm.IsRuneword {
			runewords[itm.RunewordName] = true
		}
		if _, tracked := nightmareCountessTrackedRunes[string(itm.Name)]; tracked {
			runes[string(itm.Name)] += action.GetItemQuantity(itm)
		}
	}

	return runes, runewords
}

func consumeRunewordRunes(available map[string]int, runeword item.RunewordName) bool {
	requirements := nightmareCountessRunewordRequirements[runeword]
	for runeName, quantity := range requirements {
		if available[runeName] < quantity {
			return false
		}
	}

	for runeName, quantity := range requirements {
		available[runeName] -= quantity
	}

	return true
}

func hasCountessRunewordsReady(items []data.Item, selected []item.RunewordName) bool {
	if len(selected) == 0 {
		return false
	}

	availableRunes, ownedRunewords := buildCountessInventory(items)
	for _, runeword := range selected {
		if ownedRunewords[runeword] {
			continue
		}
		if !consumeRunewordRunes(availableRunes, runeword) {
			return false
		}
	}

	return true
}

type Countess struct {
	ctx *context.Status
}

func NewCountess() *Countess {
	return &Countess{
		ctx: context.Get(),
	}
}

func (c Countess) Name() string {
	return string(config.CountessRun)
}

func (a Countess) CheckConditions(parameters *RunParameters) SequencerResult {
	farmingRun := IsFarmingRun(parameters)
	questCompleted := a.ctx.Data.Quests[quest.Act1TheForgottenTower].Completed()
	if !farmingRun && questCompleted {
		return SequencerSkip
	}

	if farmingRun && parameters != nil && parameters.SequenceSettings != nil {
		selectedRunewords := selectedCountessRunewords(parameters.SequenceSettings, a.ctx.CharacterCfg.Game.Difficulty)

		if _, isLevelingChar := a.ctx.Char.(context.LevelingCharacter); isLevelingChar &&
			a.ctx.CharacterCfg.Game.Difficulty == difficulty.Normal &&
			!a.ctx.Data.Quests[quest.Act1TheSearchForCain].Completed() {
			return SequencerSkip
		}

		if _, isLevelingChar := a.ctx.Char.(context.LevelingCharacter); isLevelingChar &&
			a.ctx.CharacterCfg.Game.Difficulty == difficulty.Normal &&
			len(selectedRunewords) > 0 {
			items := a.ctx.Data.Inventory.ByLocation(
				item.LocationInventory,
				item.LocationStash,
				item.LocationSharedStash,
				item.LocationRunesTab,
			)
			if hasCountessRunewordsReady(items, selectedRunewords) {
				return SequencerSkip
			}
		}

		if _, isLevelingChar := a.ctx.Char.(context.LevelingCharacter); isLevelingChar &&
			a.ctx.CharacterCfg.Game.Difficulty == difficulty.Nightmare &&
			len(selectedRunewords) > 0 {
			items := a.ctx.Data.Inventory.ByLocation(
				item.LocationInventory,
				item.LocationStash,
				item.LocationSharedStash,
				item.LocationRunesTab,
			)
			if hasCountessRunewordsReady(items, selectedRunewords) {
				return SequencerSkip
			}
		}
	}

	return SequencerOk
}

func (c Countess) Run(parameters *RunParameters) error {
	// Travel to boss level
	err := action.WayPoint(area.BlackMarsh)
	if err != nil {
		return err
	}

	areas := []area.ID{
		area.ForgottenTower,
		area.TowerCellarLevel1,
		area.TowerCellarLevel2,
		area.TowerCellarLevel3,
		area.TowerCellarLevel4,
		area.TowerCellarLevel5,
	}
	clearFloors := c.ctx.CharacterCfg.Game.Countess.ClearFloors

	for _, a := range areas {
		err = action.MoveToArea(a)
		if err != nil {
			return err
		}

		if clearFloors && a != area.TowerCellarLevel5 {
			if err = action.ClearCurrentLevel(false, data.MonsterAnyFilter()); err != nil {
				return err
			}
		}
	}

	err = action.MoveTo(func() (data.Position, bool) {
		areaData := c.ctx.Data.Areas[area.TowerCellarLevel5]
		countessNPC, found := areaData.NPCs.FindOne(740)
		if !found {
			return data.Position{}, false
		}

		return countessNPC.Positions[0], true
	})
	if err != nil {
		return err
	}

	// Kill Countess
	if err := c.ctx.Char.KillCountess(); err != nil {
		return err
	}

	action.ItemPickup(30)

	if clearFloors {
		return action.ClearCurrentLevel(false, data.MonsterAnyFilter())
	}
	return nil
}
