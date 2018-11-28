package static

var (
  kSelectAllJs = `
function selectAll(checkAll, className) {
  checkboxes = checkAll.form.getElementsByClassName(className);
  for (var idx = 0; checkboxes[idx]; idx++) {
    checkboxes[idx].checked = checkAll.checked;
  }
}`
)
