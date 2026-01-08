# Feature Ideas for rig.fm

A collection of potential features to enhance the terminal radio experience.

---

## Playback & Discovery

### Sleep Timer
Auto-stop playback after X minutes. Perfect for falling asleep to radio.
- Simple countdown timer
- Visual indicator showing time remaining
- Configurable default duration

### Now Playing Metadata
Show current song/artist from ICY stream metadata.
- MPV can expose ICY metadata via IPC
- Display in Now Playing panel
- Optional desktop notifications on song change

### Random Station
"I'm Feeling Lucky" - play a random station.
- Respects current filters (random rock station, random UK station, etc.)
- Quick keyboard shortcut

### Station Recommendations
"Similar to this" based on tags/genre.
- Analyze current station's tags
- Find stations with overlapping tags
- Show as suggestions in UI

---

## History & Stats

### Listening History
Track what you've played, when, and for how long.
- Store in `~/.config/rig/history.json`
- View recent history in UI
- Track total listening time per station

### Most Played
See your personal top stations ranked by play count or time.
- Separate from favorites (auto-tracked vs manual)
- Quick access to frequently played stations

### Recently Played
Quick access to last 10 stations.
- Keyboard shortcut to cycle through recent
- Shown in a dedicated section or filter

---

## Recording & Export

### Record Stream
Save current audio to MP3/WAV file.
- MPV supports stream recording
- Start/stop with keyboard shortcut
- Auto-naming with station + timestamp

### Export/Import Favorites
Share your favorites as JSON or sync across machines.
- Export to file
- Import from file
- Merge or replace options

---

## Automation

### Alarm Clock
Wake up to your favorite station at a set time.
- Set alarm time and station
- Gradual volume increase option
- Snooze functionality

### Scheduled Playback
Different stations for different times of day.
- Morning: news station
- Afternoon: music
- Evening: chill
- Cron-like scheduling

---

## Visual & UX

### Audio Visualizer
Waveform or spectrum display in terminal.
- ASCII art visualization
- Multiple visualization styles
- Toggle on/off

### Themes
Color scheme options.
- Dark (current)
- Light
- Retro/Amber
- Custom color configuration

### Mini Mode
Compact single-line view for status bars.
- Show: station name, play/pause status, volume
- Minimal footprint
- Toggle with keyboard shortcut

---

## Integrations

### Last.fm Scrobbling
Track your listening on Last.fm or ListenBrainz.
- Requires song metadata from stream
- API integration
- Privacy-conscious (opt-in)

### Desktop Notifications
System notifications when song changes.
- macOS/Linux notification support
- Configurable (on/off)
- Show album art if available

---

## Quick Access

### Quick Switch Keys
Number keys 1-9 to jump to favorite stations instantly.
- Assign stations to number slots
- Visual indicator of assigned stations
- Quick reassignment

### Station Aliases
Name your favorites with custom aliases.
- `rig play chill` - plays your "chill" station
- `rig play news` - plays your "news" station
- CLI and TUI support

---

## Priority Suggestions

**Quick Wins (Easy to implement, high value):**
1. Sleep Timer
2. Random Station
3. Recently Played

**Medium Effort:**
4. Now Playing Metadata
5. Listening History
6. Export/Import Favorites

**Larger Features:**
7. Audio Visualizer
8. Themes
9. Recording

---

## Notes

- All features should maintain the "beautiful terminal experience" philosophy
- Keep keyboard-first design
- Performance should not degrade with new features
- Consider mobile/SSH use cases (low bandwidth, no mouse)



# BUGS
- mpv seems to hog battery
- toggling off favourites view shows all stations instead of top 100
- 