// ADV cutscene subtitle dialogue. Each entry maps an (EventIndex, VoiceID)
// pair to a Text string. EventIndex identifies the cutscene, VoiceID is the
// voice line within that cutscene (determined by Ghidra analysis of the
// voice dispatch tables), and Text is the displayed subtitle.
//
// Text is split into up to 4 lines at \n. The line-splitting happens in
// splitSubtitleLines() in adv_subtitle.go; lines beyond the 4th are folded
// into the 4th. An empty Text suppresses that entry.
//
// Voice IDs are non-contiguous because event scripts skip indices  -- gaps
// represent voice lines that exist in the script but have no dialogue
// (character barks, SFX-only moments, etc.).

package text

// Pharmacy customer
var AdvSubtitles = []AdvSubtitlePatch{
	{EventIndex: 0x0007, VoiceID: 0x0000, Text: "Um, excuse me!"},
	{EventIndex: 0x0007, VoiceID: 0x0001, Text: "You dropped this."},
	{EventIndex: 0x0007, VoiceID: 0x0002, Text: "You know, you get points when\n" +
		"you buy Hoihoi-san stuff at\n" +
		"the pharmacy, and you get good\n" +
		"things by collecting them!"},
	{EventIndex: 0x0007, VoiceID: 0x0003, Text: "Man, isn't she just the best?\n" +
		"Hoihoi-san, I mean..."},
	{EventIndex: 0x0007, VoiceID: 0x0004, Text: "They're supposed to be\n" +
		"releasing all sorts of new\n" +
		"weapons and outfits soon too!\n" +
		"I can't wait!"},
	{EventIndex: 0x0007, VoiceID: 0x0005, Text: "Oh, sorry... I kind of got\n" +
		"carried away there."},
	{EventIndex: 0x0007, VoiceID: 0x0006, Text: "Well, I'll be on my way."},
	{EventIndex: 0x0007, VoiceID: 0x0007, Text: "Oh! But if you hear any news\n" +
		"about Hoihoi-san, please let\n" +
		"me know!"},

	{EventIndex: 0x0008, VoiceID: 0x0000, Text: "Oh, hey!"},
	{EventIndex: 0x0008, VoiceID: 0x0001, Text: "How's Hoihoi-san doing?"},
	{EventIndex: 0x0008, VoiceID: 0x0002, Text: "Oh, right, I just bought this!"},
	{EventIndex: 0x0008, VoiceID: 0x0003, Text: "Pretty nice, right? I can't\n" +
		"wait to get home to try it on\n" +
		"her!"},
	{EventIndex: 0x0008, VoiceID: 0x0004, Text: "Well, see ya!"},

	// Aburatsubo
	{EventIndex: 0x0009, VoiceID: 0x0000, Text: "Oh, perfect timing."},
	{EventIndex: 0x0009, VoiceID: 0x0001, Text: "Would you mind taking some of\nthese?"},
	{EventIndex: 0x0009, VoiceID: 0x0002, Text: "I redeemed my pharmacy points\n" +
		"and my prize was a \"one year\n" +
		"supply of mobile lifts\", so I\n" +
		"ended up with all this stuff."},
	{EventIndex: 0x0009, VoiceID: 0x0003, Text: "What's even the standard for a\n" +
		"one year supply?"},
	{EventIndex: 0x0009, VoiceID: 0x0004, Text: "Anyway, I heard some rumour\n" +
		"that there was a battery\n" +
		"defect and these were\n" +
		"recalled."},
	{EventIndex: 0x0009, VoiceID: 0x0005, Text: "Well, I've never seen them on\n" +
		"sale before. I guess it\n" +
		"wouldn't hurt to try them out."},
	{EventIndex: 0x0009, VoiceID: 0x0006, Text: "Anyway, use them however you\n" +
		"like. See you!"},

	{EventIndex: 0x000A, VoiceID: 0x0000, Text: "Oh, hello."},
	{EventIndex: 0x000A, VoiceID: 0x0001, Text: "Hm? Did you buy something new?"},
	{EventIndex: 0x000A, VoiceID: 0x0002, Text: "A-a?! Is that..? The rumored\n" +
		"Hoihoi-san shrine maiden\n" +
		"outfit?"},
	{EventIndex: 0x000A, VoiceID: 0x0003, Text: "I can't believe it's already\n" +
		"out... This is the greatest\n" +
		"failure of my life."},
	{EventIndex: 0x000A, VoiceID: 0x0004, Text: "No time to waste! See ya!"},
	{EventIndex: 0x000A, VoiceID: 0x0005, Text: ""},

	{EventIndex: 0x000B, VoiceID: 0x0000, Text: "Oh, hey there."},
	{EventIndex: 0x000B, VoiceID: 0x0001, Text: "Heh, I found something good\n" +
		"today."},
	{EventIndex: 0x000B, VoiceID: 0x0002, Text: "Ta-da! Magical Hoihoi-san!"},
	{EventIndex: 0x000B, VoiceID: 0x0003, Text: "It was the last one at the\n" +
		"pharmacy. I'm so glad I was\n" +
		"able to get one."},
	{EventIndex: 0x000B, VoiceID: 0x0004, Text: "Better hurry and change her\n" +
		"outfit. Hehehe."},
	{EventIndex: 0x000B, VoiceID: 0x0005, Text: "See you!"},

	{EventIndex: 0x000C, VoiceID: 0x0000, Text: "Oh, hey."},
	{EventIndex: 0x000C, VoiceID: 0x0001, Text: "Listen to this."},
	{EventIndex: 0x000C, VoiceID: 0x0002, Text: "My Hoihoi-san was chasing a\n" +
		"bug around the table and she\n" +
		"fell off."},
	{EventIndex: 0x000C, VoiceID: 0x0003, Text: "She got broken so I took her\n" +
		"in for repairs."},
	{EventIndex: 0x000C, VoiceID: 0x0004, Text: "Aaa! Hoihoi-san, she's gone..."},
	{EventIndex: 0x000C, VoiceID: 0x0005, Text: "The world is over, the bugs\n" +
		"are going to take over!"},
	{EventIndex: 0x000C, VoiceID: 0x0006, Text: ""},

	{EventIndex: 0x000D, VoiceID: 0x0000, Text: "Oh, hello."},
	{EventIndex: 0x000D, VoiceID: 0x0001, Text: "Did you hear? That new\n" +
		"product, 'Combat-san'?"},
	{EventIndex: 0x000D, VoiceID: 0x0002, Text: "Apparently if you use it at\n" +
		"the same time as Hoihoi-san,\n" +
		"it will exterminate Hoihoi-san\n" +
		"too!"},
	{EventIndex: 0x000D, VoiceID: 0x0003, Text: "I hear Mars Pharmaceutical are\n" +
		"in a huge panic over it."},
	{EventIndex: 0x000D, VoiceID: 0x0004, Text: "I'm so glad I found out before\n" +
		"getting one. Huh? What's\n" +
		"wrong?"},
	{EventIndex: 0x000D, VoiceID: 0x0005, Text: "Ah! You didn't?!"},
	{EventIndex: 0x000D, VoiceID: 0x0006, Text: "That's bad! You'd better get\n" +
		"back there quickly."},

	// kimiko finds out you live nearby.. with a hoihoi-san
	{EventIndex: 0x000E, VoiceID: 0x0000, Text: "Oh, customer! You live around\n" +
		"here?"},
	{EventIndex: 0x000E, VoiceID: 0x0001, Text: "Well, in that case, we're\n" +
		"neighbours. It's a pleasure!"},
	{EventIndex: 0x000E, VoiceID: 0x0002, Text: "A- a- uh... if you're that\n" +
		"customer, then you're using\n" +
		"Hoihoi-san around here?"},
	{EventIndex: 0x000E, VoiceID: 0x0003, Text: "Then, even right now she's..."},
	{EventIndex: 0x000E, VoiceID: 0x0004, Text: "Begone!"},
	{EventIndex: 0x000E, VoiceID: 0x0005, Text: ""},
	{EventIndex: 0x000E, VoiceID: 0x0006, Text: ""},

// Kimiko's bug infestation
	{EventIndex: 0x000F, VoiceID: 0x0000, Text: "A-ah, customer?"},
	{EventIndex: 0x000F, VoiceID: 0x0001, Text: "My room has ended up swarming\n" +
		"with bugs."},
	{EventIndex: 0x000F, VoiceID: 0x0002, Text: "I can't believe I let it get\n" +
		"like this. I'm so useless..."},
	{EventIndex: 0x000F, VoiceID: 0x0003, Text: "Eek."},
	{EventIndex: 0x000F, VoiceID: 0x0004, Text: "Please, could you use your\n" +
		"Hoihoi-san to get rid of the\n" +
		"bugs at my place?"},
	{EventIndex: 0x000F, VoiceID: 0x0005, Text: "I'm begging you, please. For\n" +
		"heaven's sake!"},

	{EventIndex: 0x0010, VoiceID: 0x0000, Text: "Huh? You were able to get them\n" +
		"all?"},
	{EventIndex: 0x0010, VoiceID: 0x0001, Text: "Ah, you saved me!"},
	{EventIndex: 0x0010, VoiceID: 0x0002, Text: "Right, I ought to thank\n" +
		"you somehow!"},
	{EventIndex: 0x0010, VoiceID: 0x0003, Text: "It's not much, but please take\n" +
		"this."},
	{EventIndex: 0x0010, VoiceID: 0x0004, Text: "Well then, I've got work to\n" +
		"do, so..."},
	{EventIndex: 0x0010, VoiceID: 0x0005, Text: "Thank you so much!"},

	{EventIndex: 0x0011, VoiceID: 0x0000, Text: "Eh? You got them already?"},
	{EventIndex: 0x0011, VoiceID: 0x0001, Text: "Ah, you saved me!"},
	{EventIndex: 0x0011, VoiceID: 0x0002, Text: "Your Hoihoi-san sure gets the\n" +
		"job done!"},
	{EventIndex: 0x0011, VoiceID: 0x0003, Text: "Oh, wait here a moment."},
	{EventIndex: 0x0011, VoiceID: 0x0004, Text: "I made this myself, but please\n" +
		"take it."},
	{EventIndex: 0x0011, VoiceID: 0x0005, Text: "Well, I've got work to do..."},
	{EventIndex: 0x0011, VoiceID: 0x0006, Text: "Thanks so much!"},
	{EventIndex: 0x0011, VoiceID: 0x0007, Text: ""},

// Kimiko and Combat-san
	{EventIndex: 0x0012, VoiceID: 0x0000, Text: "Ah, hey there, regular!"},
	{EventIndex: 0x0012, VoiceID: 0x0001, Text: "Just look at this! It's the\n" +
		"new 'Combat-san' doll. I'm\n" +
		"so glad I managed to get one!"},
	{EventIndex: 0x0012, VoiceID: 0x0002, Text: "Ah, customer... I'm sorry, but\n" +
		"our shop has already sold out."},
	{EventIndex: 0x0012, VoiceID: 0x0003, Text: "Please wait until we get more\n" +
		"in stock."},
	{EventIndex: 0x0012, VoiceID: 0x0004, Text: "Later!"},

	{EventIndex: 0x0013, VoiceID: 0x0000, Text: "Um, excuse me."},
	{EventIndex: 0x0013, VoiceID: 0x0001, Text: "Please buy this Combat-san\n" +
		"from me and don't ask any\n" +
		"questions."},
	{EventIndex: 0x0013, VoiceID: 0x0002, Text: "It was returned, but it's\n" +
		"practically brand new. It's\n" +
		"the perfect opportunity to get\n" +
		"one."},
	{EventIndex: 0x0013, VoiceID: 0x0003, Text: "Huh?! You'll take it?"},
	{EventIndex: 0x0013, VoiceID: 0x0004, Text: "Wonderful! Much obliged!"},

// Chief from Mars Pharmaceutical
	{EventIndex: 0x0014, VoiceID: 0x0000, Text: "Excuse me, young friend."},
	{EventIndex: 0x0014, VoiceID: 0x0001, Text: "I believe you just dropped\n" +
		"this."},
	{EventIndex: 0x0014, VoiceID: 0x0002, Text: "Oho! I see you're quite fond\n" +
		"of our company's 'Hoihoi-san'!"},
	{EventIndex: 0x0014, VoiceID: 0x0003, Text: "I'm here on a business trip. I\n" +
		"never expected to come across\n" +
		"such an enthusiastic user."},
	{EventIndex: 0x0014, VoiceID: 0x0004, Text: "Oh, where are my manners? As a\n" +
		"matter of fact, I am with Mars\n" +
		"Pharmaceutical－"},
	{EventIndex: 0x0014, VoiceID: 0x0005, Text: "C-Chief! What in the world\n" +
		"are you doing here?!"},
	{EventIndex: 0x0014, VoiceID: 0x0006, Text: "You'll be late to the meeting!\n" +
		"Come on, we have to hurry!"},
	{EventIndex: 0x0014, VoiceID: 0x0007, Text: "Yes, sorry."},
	{EventIndex: 0x0014, VoiceID: 0x0008, Text: "Well then, please excuse me."},

	{EventIndex: 0x0015, VoiceID: 0x0000, Text: "Hmm? Well, we meet again!"},
	{EventIndex: 0x0015, VoiceID: 0x0001, Text: "So, how is Hoihoi-san doing?"},
	{EventIndex: 0x0015, VoiceID: 0x0002, Text: "Oh, yes. I have something good\n" +
		"for you."},
	{EventIndex: 0x0015, VoiceID: 0x0003, Text: "Here, look at this."},
	{EventIndex: 0x0015, VoiceID: 0x0004, Text: "*chuckles* Equip this and\n" +
		"Hoihoi-san's performance will\n" +
		"increase several-fold."},
	{EventIndex: 0x0015, VoiceID: 0x0005, Text: "Honestly, sir... In a place\n" +
		"like this again?"},
	{EventIndex: 0x0015, VoiceID: 0x0006, Text: "You're going to be late for\n" +
		"the meeting!"},
	{EventIndex: 0x0015, VoiceID: 0x0007, Text: "Yes, sorry."},
	{EventIndex: 0x0015, VoiceID: 0x0008, Text: "Please excuse me."},
	{EventIndex: 0x0016, VoiceID: 0x0000, Text: "Oh, look who it is."},
	{EventIndex: 0x0016, VoiceID: 0x0001, Text: "How was it? That item from the\n" +
		"other day."},
	{EventIndex: 0x0016, VoiceID: 0x0002, Text: "It was actually a prototype\n" +
		"for a new product."},
	{EventIndex: 0x0016, VoiceID: 0x0003, Text: "Its performance is one thing,\n" +
		"but the design is magnificent,\n" +
		"right?"},
	{EventIndex: 0x0016, VoiceID: 0x0004, Text: "CHIEF!"},
	{EventIndex: 0x0016, VoiceID: 0x0005, Text: "Yazaki. I know what you're\n" +
		"going to say. It's a meeting,\n" +
		"right?"},
	{EventIndex: 0x0016, VoiceID: 0x0006, Text: "That's not it! You'll miss\n" +
		"your train home!"},
	{EventIndex: 0x0016, VoiceID: 0x0007, Text: "W-what?"},
	{EventIndex: 0x0016, VoiceID: 0x0008, Text: "Sorry. Please excuse me."},
	{EventIndex: 0x0016, VoiceID: 0x0009, Text: "We're going to run, Yazaki."},
	{EventIndex: 0x0016, VoiceID: 0x000A, Text: "Chief..."},
	{EventIndex: 0x0016, VoiceID: 0x000B, Text: ""},
}
