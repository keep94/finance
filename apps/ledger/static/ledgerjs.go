package static

var (
	kLedgerJs = `
function populateSelect(select, options) {
  for (var idx = 0; options[idx]; idx += 2) {
    var o = document.createElement("option");
    o.value = options[idx];
    o.text = options[idx + 1];
    select.add(o, null);
  }
}

function initAutoComplete(autoComplete) {
  autoComplete.applyLocalFilter = true;
  autoComplete.queryMatchContains = true;
  autoComplete.autoHighlight = false;
}

function Suggester(url) {
  this._url = url;
  this._suggestions = [];
}

Suggester.prototype._initRequest = function() {
  if (window.isIE) {
    return new ActiveXObject("Microsoft.XMLHTTP");
  }
  return new XMLHttpRequest();
};

Suggester.prototype.useUrl = function(url) {
  this._url = url;
  return this;
}

Suggester.prototype.getSuggestions = function() {
  if (this._url !== this._readUrl) {
    this._readUrl = this._url;
    this._suggestions = [];
    var req = this._initRequest();
    var that = this;
    req.onreadystatechange = function() {
      if (req.readyState == 4) {
        if (req.status == 200) {
          that._suggestions = JSON.parse(req.responseText);
        }
      }
    };
    req.open("GET", this._url, true);
    req.send(null);
  }
  return this._suggestions;
};`
)
