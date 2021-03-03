window.onclick = function() {
  document
    .getElementsByClassName("dropdown-menu")[0]
    .classList.remove("active");
};

document.getElementsByClassName("my-menu-link")[0].onclick = function(e) {
  e.stopPropagation();
  this.nextElementSibling.classList.toggle("active");
};

document.querySelectorAll(".dropdown-menu ul li")[0].onclick = function(e) {
  e.stopPropagation();
  const url = this.getAttribute("go");

  window.open(url, "_self");
};
