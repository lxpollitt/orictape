# orictape
Found some decaying cassette tapes with all the programs you wrote as kid on them?  You'll need a command line tool to show you the mangled wave forms, where the suspect bits are, and give you the best chance of reconstructing your childhood masterpieces.

## About Oric
![Oric 1](/img/Oric1.png)

Oric was the name used by Tangerine Computer Systems for a series of home computers, including the original *Oric-1*, its successor the *Oric Atmos* and the later *Oric Stratos/IQ164* and *Oric Telestrat* models.

Released in 1983, the Oric-1 was based on a MOS 1 MHz 6502A CPU, and came in 16 KB or 48 KB RAM variants for £129 and £169 respectively. Both versions had a 16 KB ROM containing the operating system and a modified BASIC interpreter.  During 1983, around 160,000 Oric-1 computers were sold in the UK, plus another 50,000 in France.

## About the tool
Tool shows:
* Audio audio waveform with interpretation of bits, highlighting the bits where the audio is damaged.
* Each corresponding byte, highlighting bytes where audio was damaged, check sum errors, and unrecognized symbols.
* The program itself in Basic, again highlighting the suspect bits.

Scroll around in any direction using the cursor keys.

![Screen Shot](/img/screenshot1.png)


## Emulators
Once you've reconstructed your programs, you'll need something to run them on. Here's a few to try:
* http://www.bannister.org/software/oric.htm
* http://www.emucamp.com/oric.html

### TO DO

- [ ] Make the code not look like the first golang program anyone ever wrote.
- [ ] Compare two copies of the tape and take the best bits from each.
- [ ] Ability to edit bits, bytes, and keywords.
- [ ] Export reconstructed program as `.tap` and `.wav` files.
