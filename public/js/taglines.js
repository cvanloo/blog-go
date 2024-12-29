const spanOpening = `<span class="weird">`;
const spanClosing = `</span>`;
class TaglinesCached {
	run() {
		const ts = JSON.parse(localStorage.getItem('taglines'));
		if (!Array.isArray(ts) || ts.length == 0) {
			throw "empty (cached) taglines";
		}
		let i = +localStorage.getItem('taglines_index') + 1;
		if (i >= ts.length) { i = 0 }
		localStorage.setItem('taglines_index', i);
		return ts[i];
	}
}
class TaglinesFresh {
	constructor(text) {
		this.text = text
	}
	run() {
		const lines = this.splitLines(this.text);
		this.guardEmptyLines(lines);
		const taglines = lines.map(this.strToTagline);
		this.shuffle(taglines);
		return this.cache(taglines);
	}
	splitLines(text) {
		return text.trim().split('\n');
	}
	guardEmptyLines(lines) {
		if (lines.length === 0 || (lines.length === 1 && lines[0].trim() === "")) {
			throw "empty taglines";
		}
	}
	strToTagline(str) {
		let tagline = "…";
		let lastLoc = 0;
		for (const weirdLoc of str.trim().matchAll(/<|>/g)) {
			tagline += str.substring(lastLoc, weirdLoc.index);
			if (weirdLoc[0] === "<") {
				tagline += spanOpening;
			} else if (weirdLoc[0] === ">") {
				tagline += spanClosing
			}
			lastLoc = weirdLoc.index + 1;
		}
		tagline += str.substring(lastLoc);
		tagline += ")";
		return tagline;
	}
	shuffle(taglines) {
		let c = taglines.length;
		while (c > 0) {
			let r = Math.floor(Math.random() * c);
			c--;
			[taglines[c], taglines[r]] = [taglines[r], taglines[c]];
		}
		return taglines;
	}
	cache(taglines) {
		localStorage.setItem('taglines', JSON.stringify(taglines));
		localStorage.setItem('taglines_index', 0);
		return taglines[0];
	}
}
async function checkVersion(resp) {
	const hs = resp.headers;
	if (hs.has('Last-Modified')) {
		const version = new Date(hs.get('Last-Modified'));
		const cachedVersion = new Date(localStorage.getItem('taglines_version'));
		if (version <= cachedVersion) {
			return new TaglinesCached();
		}
		localStorage.setItem('taglines_version', version);
	}
	return new TaglinesFresh(await resp.text());
}
const t = document.getElementById('tagline');
fetch('/taglines')
	.then(checkVersion)
	.then(store => store.run())
	.then(tagline => t.innerHTML = tagline)
	.catch(console.error)
