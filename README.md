# Evenicle Unburger Mod

An English re-translation and restoration mod for the MangaGamer PC release of **Evenicle 1.04**.

## Why this mod exists

When I first sat down to fix Evenicle, I expected something roughly on the scale of my Dohna Dohna project. Dohna Dohna had plenty of localization choices I wanted to undo, missing honorifics, terminology I wanted to restore, and a fair number of lines that needed another translation, but most of the English script was still usable.

I expected Evenicle to be the same kind of job: restore the obvious cultural material, fix the more famous bad lines, clean up the terminology, and move on. It became clear fairly quickly that it was going to be something much larger.

The spectacular mistakes are the easiest part to demonstrate because some lines simply reverse the meaning, drop information, replace a Japanese joke with an unrelated Western reference, or turn a relationship term into a meme. There are even named objects and places that were changed into something with no real connection to the Japanese at all.

What took longer to notice was how much of the damage came from edits that look minor when you meet them one at a time. A changed pronoun or missing relationship marker rarely looks dramatic on its own, and neither does a quiet reaction made louder, a recurring speech quirk normalized, a small extra insult, or a cultural reference replaced with a generic explanation. Repeated across a scene, though, those choices start changing the tone; repeated across the whole game, they can leave a character more vulgar, more flippant, less affectionate, or simply less like the person in the Japanese script.

Once I started seeing that pattern, the modest correction mod I had planned stopped making sense. I began checking the English release against the Japanese script line by line. Large stretches needed to be translated again, other lines only needed a word or two changed, and plenty of official lines were already fine enough to stay exactly as they were.

So while this project started as a localization cleanup, the finished mod is much closer to a re-translation than a small fix pack.

It is still not a word-for-word translation. I do not copy Japanese sentence structure into English, and I do not replace a good official line just to make it different. I want the English to read normally while keeping what was actually written in Japanese: the meaning, characterization, relationship language, jokes, cultural references and concrete details.

## How far the official localization went

The official script ranges from straightforward mistranslations to deliberate rewrites, and some of the stranger choices can even look funny in isolation. They stop looking harmless once the same editorial habits keep turning up throughout the game. A few of the easier examples to summarize are:

- `三、四十枚` calls for **thirty or forty** giant leaves. The official English says **three or four**.
- Towa's **"Don't fight over me!"** became **"Do it for me!"**, reversing the point of the line.
- One explanation reverses the setting's history and makes the **world managers create Mother Eve**, when the Japanese says Mother Eve created them.
- The Outlaw-blessing discussion says their suffering comes **not from Mother Eve abandoning them, but from punishment**. The English changes that into abandonment **and** punishment.
- Arthur can move on her own after the paralysis spell is removed. The English says she is still unable to move.
- One sexual scene reverses who is doing what to whom and makes **Aster** perform an action that **Riche** performs in Japanese.
- **Senpai** was replaced with **Chad**, including the speaker label.
- **Onee-sama** became **Mistress**, turning a reverent older-sister address into master/servant language.
- **Ramen** became pizza in one scene and tacos in another, while the following **charumera** ramen-vendor joke was reduced to a generic bell.
- **Nakamuraya croquettes** became pizza or generic snacks. The typhoon/croquette joke disappeared with them.
- **Onigiri and miso soup** became sandwiches and soup.
- A town known for **alcohol, miso and dried bonito** became a town known for beer, yogurt and cheese.
- The classic **Manju Kowai** rakugo reference was rebuilt around **cake pops**.
- The recurring Alicesoft medicine **Seirogan** became **Ibeprofun** in dialogue, item text and battle use.
- **Hyakuen** became **Bigbux**, and the English script invented a lumberjack named Bigbux to support the new name. The Japanese name is tied to the idea of forming **"a hundred bonds"**.
- **High Tension Island** became **Fervor/Frenzy**.
- A stock complaint about **chronic stomach pains** became **chronic hemorrhoids**.
- **"You cannot escape from the Great Demon King"** was replaced with a **Babadook** joke.
- Towa's father, **Rare Cheese**, was renamed **Laertes**, replacing a food-name joke with a Shakespeare reference.
- **"I'll get married when I get home"** became the unrelated **"one day from retirement"** cliché.
- **Sea of Corruption**, a Nausica reference, became the generic **sea of filth**.
- **Daidarabotchi** became a generic titan, while **Nopperabou** and **Oitekebori** were flattened into generic descriptions.
- **Shogi, hanafuda, Iigo and beigoma** became chess, generic cards, Othello and generic tops.
- A **married-couple manzai** joke became a generic comedy-duo joke.
- **Herbivore/carnivore** dating types became **beta/alpha**.
- **Yutori generation** became **millennial**.
- **Miko** was repeatedly replaced with things like maiden or prophet.
- **Yuri**, **BL**, **NTR**, **chuunibyou**, **kabedon**, **ero flag** and **lucky sukebe** were often replaced with one-off explanations or unrelated English phrasing instead of keeping the terms being discussed.
- The fictional **Ropper** computer became a real **X68000**.
- **Loli-only Hanny** became **Pedo Hanny**.
- **Kamishibai** became generic story/storytime/pictures wording.
- The recurring Alicesoft **Hyahha** cry became **Yahoo/Wahoo**.
- **Japanese-style** became **eastern-influenced**.
- Rocinante's deliberately odd `ミー / ユー` speech was turned into a recurring French gimmick with *bonjour, merci, oui, moi, magnifique* and *sacre bleu*. The French is not in his Japanese dialogue.

Those are the conspicuous cases. Much more of the work was spent on things that are harder to turn into a screenshot: getting the subject and object of a sentence back in the right places, keeping recurring terminology consistent, restoring physical details or motives that had disappeared, and repairing setups that had been changed so heavily that the original joke no longer made sense. Characterization was often affected in exactly that slower way, through a long run of smaller edits rather than one spectacularly wrong line.

A much longer selection is collected in [docs/NOTABLE_CHANGES.md](docs/NOTABLE_CHANGES.md).

## Character voices and relationships

Some of the changes that bothered me most were not single mistranslations at all. They were recurring edits that slowly changed how somebody sounded.

Kinou and Sora deliberately refer to themselves by name in parts of the Japanese script. The official English often normalizes those lines to ordinary *I/me*, which removes a noticeable part of their voice. The mod restores the self-reference where the Japanese actually uses it.

Millet has the same pattern during her temporary childlike regression, while Beauty has a much smaller state-specific version of it. Erimo also occasionally leans into an older-woman persona by calling herself **Onee-san**. None of these are blanket replacements; they are kept only in the scenes and states where the source uses them.

For relationship language, I keep **Onii-chan, Onee-chan, Onee-sama, Aniue, Aneue, -san, -chan, -kun, -sama, -dono, -sensei** and other forms when they are actually part of how one character addresses another. Descriptive relationships still use ordinary English when that is what the sentence calls for, and names follow Japanese order.

Honorifics are also more than small markers of politeness or characterization. They are part of the cultural framework these relationships are written in, and many familiar Japanese character and relationship tropes are built around who uses **-san**, **-chan**, **-sama**, **senpai**, **Onee-sama** and similar forms with whom. Remove that layer and some of those dynamics lose their identity even when the basic plot information survives. Combined with flattened self-reference, changed nicknames and a rougher or more casual register, the result can still tell roughly the same story while making the people in it feel noticeably different.

I also keep Japanese cultural and genre terms such as **miko, yuri, BL, chuunibyou, kabedon, NTR, lucky sukebe, kamishibai** and **senpai** instead of routinely replacing them with explanations or Western analogues. Food, games, folklore and pop-culture references stay specific when that specificity is part of the scene or joke.

## What the mod covers

I ended up reviewing player-facing English throughout the game rather than stopping with the main story. That includes:

- main story dialogue and narration;
- side scenes, quests and events;
- battle messages and battle helper text;
- menus and UI strings;
- skill and item names and descriptions;
- location text;
- quest and request messages;
- system/helper text that is actually shown to the player;
- other English text surfaces found during the full script review.

Good official English stayed when it already matched the Japanese and read naturally. I was never interested in changing a line just to increase the size of the patch; if the retail wording worked, it stayed.

## How long this turned into

I started working on Evenicle in **February 2026** and kept coming back to it through **October 2026**, usually in bursts alongside my other projects. By the time the full review was finished, it had grown far beyond the job I thought I was taking on when I first opened the script.

At the beginning I was thinking in terms of a list: restore these terms, fix these scenes, undo these known localization choices. The deeper I went, the less useful that list became. Eventually the question was no longer "which famous changes should I undo?" It was simply: **does this line say what the Japanese says, and does this still sound like the same character?**

That led to a complete line-by-line review of the player-facing script, followed by another read focused on whether the corrected English still sounded natural in context. It was much more work than I planned for Evenicle, but once the scale of the problem was clear, doing only the famous examples would have left too much behind.

## Who this mod is for

This mod is for people who want to play the English PC release while staying much closer to the original Japanese script. If you like the MangaGamer localization exactly as it is, that version already exists; this version keeps the English readable while undoing rewrites that changed meaning, characterization, relationship language, jokes, cultural material or recurring terminology.

Colorful official lines that already communicate the Japanese naturally stay. I am not trying to make the script stiff or mechanical, only to remove rewrites that pulled it away from the source.

## Current state

The full script review is finished, and I have been playing the patched game myself since the final changes without running into anything broken in the parts I checked. A game this large can still hide a typo, an awkward line or a layout problem, so concrete in-game reports are welcome.

The broad translation work is done. If something turns up later, I would rather fix that specific problem than reopen the whole script again.

## Installation

The mod is for the **MangaGamer English PC release, version 1.04**.

### Installing a release ZIP

1. Back up `Evenicle.ain`, `Evenicle.ex`, `EvenicleCG2.afa` and `Version.txt` from the game directory.
2. Extract the release ZIP into the main Evenicle folder.
3. Allow those files to be overwritten.

To uninstall the mod, restore the original files.

### Building and installing from source

The recommended path is the interactive TUI. It configures the game and AliceTools paths, captures and validates the pristine English 1.04 files, builds the patch, and can install it without manually editing `.local/config.env`.

Linux:

```bash
./tui.sh
```

Windows:

```bat
tui.cmd
```

On the first run choose **Setup / reconfigure**. After setup, choose **Build + install patch**. The same menu can later build without installing, recapture the pristine base files, clean generated output, or change the configuration.

For command-line use, after the one-time capture/configuration step you can build and install in one command:

```bash
./build.sh install
```

```bat
build.cmd install
```

The installer is guarded: it accepts the supported pristine game state or a previously installed build with verified `.unburger-backup` files, and it always rebuilds from the captured pristine base.

## Building from source

The repository contains the translation data and build tools, not the game itself. Building requires files from your own copy of the MangaGamer English PC release, version 1.04.

**Recommended:** use `./tui.sh` on Linux or `tui.cmd` on Windows. Choose **Setup / reconfigure** once, then **Build patch** to create the release ZIP or **Build + install patch** to build and install it directly.

The command-line builder remains available for automation. See [docs/BUILDING.md](docs/BUILDING.md) for requirements, the complete TUI workflow, CLI equivalents and deterministic/byte-identity details.

## Documentation

- [Notable translation changes](docs/NOTABLE_CHANGES.md)
- [Building from source](docs/BUILDING.md)

## Credits

[AliceTools](https://github.com/nunuhara/alice-tools) by nunuhara is used for Alicesoft archive and script handling.

This is an unofficial fan mod. The repository contains patch data and tooling only; no game files are distributed here.
