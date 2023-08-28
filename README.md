# Used this for like a week. 🫡

# (Was) Backend for my blog @ [lolwierd.ml](lolwierd.com)

Uses notion as a CMS for blogs.

Blog frontend is deployed using hugo.

This backend basically fetches posts (with updated property set to true) from a notion database, converts them into markdown and saves the corresponding markdown to the hugo content/posts folder. Hugo can then pick those files up and do its thing.

Direct any and all questions @ me@lolwierd.com or @lolwierrd on twitter.

> I will add the usage instructions shortly

## Known Bugs / Missing Features
- [x] Cannot rename slug (does not delete original post after rename)
- [ ] Cannot delete posts from notion
- [x] Set updated to false after fetching and building
- [ ] Add embed support for YT, gists, vimeo, tweets.
- [ ] Add callout support
- [ ] Add more syntax highlighters to original theme
- [ ] Add ToDo support to original theme
- [ ] Add color support (??)
