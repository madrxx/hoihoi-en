// Mission text patches  -- titles, descriptions, and objectives for all
// 20 mission records in the mission BSV table.
//
// Each patch has:
//   - Title, Description: replace the corresponding string; empty = keep original.
//   - ConfirmLines: replace the confirmation/objective lines shown before
//     starting a mission. This is a 4-element slice (one line per display
//     slot). Empty strings suppress that line. nil ConfirmLines leaves all
//     existing text untouched.
//
// Missions 0-3 are Area 1 (1-1 through 1-4), 4-9 are Area 2, etc.

package text

var Missions = []MissionPatch{
	{
		Index:       0,
		Title:       "My First Job",
		Description: "Kill 10 cockroaches.",
		ConfirmLines: []string{
			"1-1",
			"Kill 10",
			"cockroaches.",
			"",
		},
	},
	{
		Index:       1,
		Title:       "Slow and Sneaky",
		Description: "Kill 14 cockroaches.",
		ConfirmLines: []string{
			"1-2",
			"Kill 14",
			"cockroaches.",
			"",
		},
	},
	{
		Index:       2,
		Title:       "Tabletop Battle",
		Description: "Kill 20 pests.",
		ConfirmLines: []string{
			"1-3",
			"Kill 20 pests.",
			"",
			"",
		},
	},
	{
		Index:       3,
		Title:       "Don't Follow Along",
		Description: "Kill the centipede.",
		ConfirmLines: []string{
			"1-4",
			"Kill the",
			"centipede.",
			"",
		},
// Area 2 (missions 4-9)
	},
	{
		Index:       4,
		Title:       "Insect Air Force",
		Description: "Kill 26 pests.",
		ConfirmLines: []string{
			"2-1",
			"Kill 26 pests.",
			"",
			"",
		},
	},
	{
		Index:       5,
		Title:       "Attic Attack",
		Description: "Kill 35 termites.",
		ConfirmLines: []string{
			"2-2",
			"Kill 35",
			"termites.",
			"",
		},
	},
	{
		Index:       6,
		Title:       "Cake Defense Force",
		Description: "Protect the cake and kill 50 pests.",
		ConfirmLines: []string{
			"2-3",
			"Protect the",
			"cake and kill",
			"50 pests.",
		},
	},
	{
		Index:       7,
		Title:       "Kitchen Panic",
		Description: "Kill 34 cockroaches.",
		ConfirmLines: []string{
			"2-4",
			"Kill 34",
			"cockroaches.",
			"",
		},
	},
	{
		Index:       8,
		Title:       "Two Hundred Feet",
		Description: "Kill 2 centipedes.",
		ConfirmLines: []string{
			"2-5",
			"Kill 2",
			"centipedes.",
			"",
		},
	},
	{
		Index:       9,
		Title:       "Operation Kimiko",
		Description: "Kill 60 pests.",
		ConfirmLines: []string{
			"2-6",
			"Kill 60 pests.",
			"",
			"",
		},
	},
// Area 3 (missions 10-15)
	{
		Index:       10,
		Title:       "Night Shift",
		Description: "Kill 40 cockroaches.",
		ConfirmLines: []string{
			"3-1",
			"Kill 40",
			"cockroaches.",
			"",
		},
	},
	{
		Index:       11,
		Title:       "Cake Defense II",
		Description: "Protect the cake and kill 60 pests.",
		ConfirmLines: []string{
			"3-2",
			"Protect the",
			"cake and kill",
			"60 pests.",
		},
	},
	{
		Index:       12,
		Title:       "Queen of the Attic",
		Description: "Kill the queen ant.",
		ConfirmLines: []string{
			"3-3",
			"Kill the queen",
			"ant.",
			"",
		},
	},
	{
		Index:       13,
		Title:       "Let's Exterminate!",
		Description: "Kill 25 camel crickets.",
		ConfirmLines: []string{
			"3-4",
			"Kill 25 camel",
			"crickets.",
			"",
		},
	},
	{
		Index:       14,
		Title:       "Danger Pattern Den",
		Description: "Destroy the wasp nest.",
		ConfirmLines: []string{
			"3-5",
			"Destroy the",
			"wasp nest.",
			"",
		},
	},
	{
		Index:       15,
		Title:       "All Monsters Attack",
		Description: "Kill 80 insects.",
		ConfirmLines: []string{
			"3-6",
			"Kill 80",
			"insects.",
			"",
		},
	},
	{
// Area 4 (mission 16)
		Index:       16,
		Title:       "Quick & Destroy",
		Description: "? ? ? ? ?",
		ConfirmLines: []string{
			"4-1",
			"Defeat",
			"Combat-san.",
			"",
		},
	},
	{
		Index:       17,
// Special missions (17-19)
		Title:       "300-Bug Slash!",
		Description: "Kill 300 pests.",
		ConfirmLines: []string{
			"SP1",
			"Kill 300",
			"pests.",
			"",
		},
	},
	{
		Index:       18,
		Title:       "Double Generator",
		Description: "Kill the queen ant and destroy the wasp nest.",
		ConfirmLines: []string{
			"SP2",
			"Kill the queen",
			"ant & destroy",
			"the wasp nest.",
		},
	},
	{
		Index:       19,
		Title:       "RRX Counterattack",
		Description: "Defeat Combat-san.",
		ConfirmLines: []string{
			"SP3",
			"Defeat",
			"Combat-san.",
			"",
		},
	},
}
