// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
)

var favicon []byte

func init() {
	var err error
	if favicon, err = base64.StdEncoding.DecodeString(faviconB64); err != nil {
		panic(err)
	}
}

var rootStyle = `
h1, h2, h3, h4, h5, h6 {
	margin-bottom: 0.2em;
	margin-top: 0.2em;
}
.err {
	background: #F44;
	border: 1px solid #888;
	border-radius: 10px;
	padding: 10px;
	display: none;
}
#periphExtra-section {
	margin-bottom: 1rem;
}
`

var rootJs = `
function post(url, data, callback) {
	let hdr = {
		body: JSON.stringify(data),
		credentials: "same-origin",
		headers: {"Content-Type": "application/json; charset=utf-8"},
		method: "POST",
	};
	fetch(url, hdr).then(checkStatus).then(callback).catch(err => onError(url, err));
}

function checkStatus(res) {
  if (res.status >= 200 && res.status < 300) {
    return res.json();
  }
	var err = new Error(res.statusText);
	err.response = res;
	throw err;
}

function onError(url, err) {
	let e = document.getElementById("err");
	if (e.innerText) {
		e.innerText = e.innerText + "\n";
	}
	e.innerText = e.innerText + url + ": " + err.toString() + "\n";
	e.style.display = "block";
}

window.onload = function() {
	refreshHeader();
	refreshI2C();
	refreshSPI();
	refreshState();
};

function refreshHeader() {
	post("/api/header/_all", {}, function(res) {
		let root = document.getElementById("section-gpio");
		for (var key in res) {
			let e = root.appendChild(document.createElement("header-elem"));
			e.setup(key, res[key].Pins);
		}
	});
}

function refreshI2C() {
	post("/api/i2c/_all", {}, function(res) {
		let root = document.getElementById("section-i2c");
		for (let i = 0; i < res.length; i++) {
			let e = root.appendChild(document.createElement("i2c-elem"));
			e.setup(res[i].Name, res[i].Number, res[i].Err, res[i].SCL, res[i].SDA);
		}
	});
}

function refreshSPI() {
	post("/api/spi/_all", {}, function(res) {
		let root = document.getElementById("section-spi");
		for (let i = 0; i < res.length; i++) {
			let e = root.appendChild(document.createElement("spi-elem"));
			e.setup(res[i].Name, res[i].Number, res[i].Err, res[i].CLK, res[i].MOSI, res[i].MISO, res[i].CS);
		}
	});
}

function refreshState() {
	post("/api/server/state", {}, function(res) {
		document.title = "periph-web - " + res.Hostname;
		document.getElementById("periphExtra").innerText = res.PeriphExtra;
		let root = document.getElementById("section-drivers-loaded");
		if (!res.State.Loaded.length) {
			root.display = "hidden";
		} else {
			root.setup(["Drivers loaded"]);
			for (var i = 0; i < res.State.Loaded.length; i++) {
				root.appendRow([res.State.Loaded[i]]);
			}
		}
		root = document.getElementById("section-drivers-skipped");
		if (!res.State.Skipped.length) {
			root.display = "hidden";
		} else {
			root.setup(["Drivers skipped", "Reason"]);
			for (var i = 0; i < res.State.Skipped.length; i++) {
				root.appendRow([res.State.Skipped[i].D, res.State.Skipped[i].Err]);
			}
		}
		root = document.getElementById("section-drivers-failed");
		if (!res.State.Failed.length) {
			root.display = "hidden";
		} else {
			root.setup(["Drivers failed", "Error"]);
			for (var i = 0; i < res.State.Failed.length; i++) {
				root.appendRow([res.State.Failed[i].D, res.State.Failed[i].Err]);
			}
		}
	});
}
`

var customElems = `
<template id="template-data-table-elem">
	<style>
		th {
			background-color: #4CAF50;
			color: white;
		}
		th, td {
			padding: 0.5rem;
			border-bottom: 1px solid #ddd;
		}
		tr:hover {
			background-color: #CCC;
		}
		tr:nth-child(even):not(:hover) {
			background: #f5f5f5;
		}
		table {
			margin-bottom: 1rem;
		}
	</style>
	<div>
		<table>
		<thead></thead>
		<tbody></tbody>
		</table>
	</div>
</template>
<script>
(function() {
  let tmpl = document.querySelector("#template-data-table-elem");
	window.customElements.define("data-table-elem", class extends HTMLElement {
		constructor() {super(); this.attachShadow({mode: "open"}).appendChild(tmpl.content.cloneNode(true));}
		setup(hdr) {
			let root = this.shadowRoot.querySelector("thead");
			for (let i = 0; i < hdr.length; i++) {
				root.appendChild(document.createElement("th")).innerText = hdr[i];
			}
		}
		appendRow(line) {
			let tr = this.shadowRoot.querySelector("tbody").appendChild(document.createElement("tr"));
			let items = [];
			for (let i = 0; i < line.length; i++) {
				let cell = tr.appendChild(document.createElement("td"));
				if (line[i] instanceof Element) {
					cell.appendChild(line[i]);
					items[i] = line[i];
				} else {
					cell.innerText = line[i];
					items[i] = cell;
				}
			}
			return items;
		}
	});
}());
</script>

<template id="template-drivers-elem">
	<data-table-elem></data-table-elem>
</template>
<script>
(function() {
  let tmpl = document.querySelector("#template-drivers-elem");
	window.customElements.define("drivers-elem", class extends HTMLElement {
		constructor() {super(); this.attachShadow({mode: "open"}).appendChild(tmpl.content.cloneNode(true));}
		setup(hdr) {
			this.shadowRoot.querySelector("data-table-elem").setup(hdr);
		}
		appendRow(row) {
			this.shadowRoot.querySelector("data-table-elem").appendRow(row);
		}
	});
}());
</script>

<template id="template-gpio-elem">
	<style>
	div {
		background: #FCF;
		border: 1px solid #888;
		border-radius: 10px;
		padding: 10px;
	}
	</style>
	<div>
		<input type="checkbox" id="state">
		<label for="state"></label>
	</div>
</template>
<script>
(function() {
  let tmpl = document.querySelector("#template-gpio-elem");
	window.customElements.define("gpio-elem", class extends HTMLElement {
		constructor() {super(); this.attachShadow({mode: "open"}).appendChild(tmpl.content.cloneNode(true));}
		connectedCallback() {
			let l = this.shadowRoot.querySelector("label");
			l.addEventListener("click", function(e) {
				//e.preventDefault();
				alert(e);
			});
		}
		setup(name, func) {
			this.shadowRoot.querySelector("label").textContent = name + "/" + func;
		}
	});
}());
</script>

<template id="template-header-elem">
	<data-table-elem></data-table-elem>
</template>
<script>
(function() {
  let tmpl = document.querySelector("#template-header-elem");
	window.customElements.define("header-elem", class extends HTMLElement {
		constructor() {super(); this.attachShadow({mode: "open"}).appendChild(tmpl.content.cloneNode(true));}
		setup(name, pins) {
			let data = this.shadowRoot.querySelector("data-table-elem");
			let cols = 1;
			if (pins) {
				cols = pins[0].length;
			}
			let hdr = [name];
			for (let i = 1; i < cols; i++) {
				hdr[i] = "";
			}
			data.setup(hdr);
			for (let y = 0; y < pins.length; y++) {
				let row = pins[y];
				let items = [];
				for (let x = 0; x < row.length; x++) {
					items[x] = document.createElement("gpio-elem");
				}
				items = data.appendRow(items);
				for (let x = 0; x < items.length; x++) {
					items[x].setup(row[x].Name, row[x].Func);
				}
			}
		}
	});
}());
</script>

<template id="template-i2c-elem">
	<data-table-elem></data-table-elem>
</template>
<script>
(function() {
  let tmpl = document.querySelector("#template-i2c-elem");
	window.customElements.define("i2c-elem", class extends HTMLElement {
		constructor() { super(); this.attachShadow({mode: "open"}).appendChild(tmpl.content.cloneNode(true)); }
		setup(name, number, err, scl, sda) {
			let data = this.shadowRoot.querySelector("data-table-elem");
			data.setup([name, ""]);
			if (number != -1) {
				data.appendRow(["Number", number]);
			}
			if (err) {
				data.appendRow(["Error", err]);
			}
			if (scl) {
				data.appendRow(["SCL", scl]);
			}
			if (sda) {
				data.appendRow(["SDA", sda]);
			}
		}
	});
}());
</script>

<template id="template-spi-elem">
	<data-table-elem></data-table-elem>
</template>
<script>
(function() {
  let tmpl = document.querySelector("#template-spi-elem");
	window.customElements.define("spi-elem", class extends HTMLElement {
		constructor() {super(); this.attachShadow({mode: "open"}).appendChild(tmpl.content.cloneNode(true));}
		setup(name, number, err, clk, mosi, miso, cs) {
			let data = this.shadowRoot.querySelector("data-table-elem");
			data.setup([name, ""]);
			if (number != -1) {
				data.appendRow(["Number", number]);
			}
			if (err) {
				data.appendRow(["Error", err]);
			}
			if (clk) {
				data.appendRow(["CLK", clk]);
			}
			if (mosi) {
				data.appendRow(["MOSI", mosi]);
			}
			if (mosi) {
				data.appendRow(["MISO", miso]);
			}
			if (cs) {
				data.appendRow(["CS", cs]);
			}
		}
	});
}());
</script>
`

var rootPage = []byte(`<!DOCTYPE html>
<meta charset="utf-8" />
<title>periph-web</title>
<style>` + rootStyle + `</style>
<script>` + rootJs + `</script>
` + customElems + `
<div class="err" id="err"></div>
<div id="section-state">
	<h1>periph's state</h1>
	<div id="periphExtra-section">
		Using <strong>periph.io/x/extra</strong>: <span id="periphExtra"></span>
	</div>
	<div>
		<drivers-elem id="section-drivers-loaded"></drivers-elem>
		<drivers-elem id="section-drivers-skipped"></drivers-elem>
		<drivers-elem id="section-drivers-failed"></drivers-elem>
	</div>
</div>
<h1>GPIO</h1>
<div id="section-gpio"></div>
<h1>I²C</h1>
<div id="section-i2c"></div>
<h1>SPI</h1>
<div id="section-spi"></div>
`)

// Created with:
// python -c "import base64;a=base64.b64encode(open('periph-web.png','rb').read()); print '\n'.join(a[i:i+70] for i in range(0,len(a),70))"
const faviconB64 = "" +
	"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" +
	"AAAElFTkSuQmCC"
