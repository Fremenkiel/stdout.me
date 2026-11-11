CodeMirror.prototype.getSelectedText = function () {
  const selected = this.doc.getSelection();

  if (selected.length > 0) {
    const pos = this.doc.getCursor("from");
    return {
      value: selected,
      lineNumber: pos.line,
      startChar: pos.ch,
    };
  }

  const value = this.doc.getValue();
  return {
    value,
    lineNumber: 0,
    startChar: 0,
  };
}

const editor = CodeMirror.fromTextArea(document.getElementById('textarea-query'), {
  lineNumbers: true,
  mode: 'text/x-sql',
  keyMap: 'vim',
});

document.addEventListener('keydown', (e) => {
  if (e.key === "Enter" &&
    (e.metaKey || e.ctrlKey)) {
    document.getElementById('submit-query').click();
  }
});

function RenderErrorSnippet(id, lineNumber, startChar, endChar) {
  const el = document.getElementById(id);

  // remove old snippets
  ResetErrorSnippet(el);

  // if line number is not set, don't show snippet
  if (lineNumber === 0) {
    return;
  }

  const selected = editor.getSelectedText();

  if (selected.lineNumber !== 0) {
    lineNumber += selected.lineNumber;
  }

  if (selected.startChar !== 0) {
    startChar += selected.startChar;
    endChar += selected.startChar;
  }

  const lineIndex = lineNumber - 1;
  const firstLineIndex = lineIndex == 0 ? 0 : lineIndex - 1;
  const lastLineIndex = lineIndex + 1;

  const firstLineNumber = lineNumber == 1 ? 1 : lineNumber - 1;

  // get line value -+ 1
  const value = editor.doc.getRange({line: firstLineIndex, ch: 0}, {line: lastLineIndex});

  // mark editor text as error
  editor.doc.markText(
    { line: lineIndex, ch: startChar},
    { line: lineIndex, ch: endChar},
    { className: "text-error", clearOnEnter: true }
  );

  // create error snippet
  var snippet = CodeMirror(el, {
    value: value,
    firstLineNumber: firstLineNumber,
    lineNumbers: true,
    mode: 'text/x-sql',
    readOnly: true,
    showCursorWhenSelecting: false,
  });

  // mark snippet text as error
  snippet.doc.markText(
    { line: 1, ch: startChar},
    { line: 1, ch: endChar},
    { className: "text-error" }
  );

  // mark line number as error
  snippet.doc.addLineClass(1, "wrap", "error-line")

  // resize editor to line height
  el.querySelector(".CodeMirror").style["height"] = 
    el.querySelector(".CodeMirror-sizer").style["min-height"];
}

function ResetErrorSnippet(el) {
  for(var i = 0; i < el.childNodes.length; i++) {
    el.childNodes[i].remove();
  }
}

function SwitchKeyMap(element) {
  const isVim = element.value === "true";

  editor.setOption("keyMap", isVim ? "default" : "vim");

  element.value = String(!isVim);
}
