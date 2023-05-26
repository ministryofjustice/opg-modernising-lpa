/**
* A modified adaptation of the Design System header script
* To initialise the One Login header, run:
* new window.CrossServiceHeader(document.querySelector("[data-module='one-login-header']")).init();
*/
export function CrossServiceHeader ($module) {
  this.$header = $module
  this.$navigation = $module.querySelectorAll("[data-one-login-header-nav]")
  this.$header.classList.add('js-enabled')
}

/**
* Initialise header
*
* Check for the presence of the header, menu and menu button – if any are
* missing then there's nothing to do so return early.
*/
CrossServiceHeader.prototype.init = function () {
  if (!this.$header && !this.$navigation.length) {
    return
  }
  /**
  * The header can render with one or two navigation elements which collapse 
  * into dropdowns on the mobile variation. This initialises the dropdown 
  * functionality for all navs that have a menu button which has: 
  * 1. a class of .js-x-header-toggle   
  * 2. an aria-controls attribute which can be mapped to the ID of the element
  * that should be hidden on mobile
  */ 
  for (var i = 0; i < this.$navigation.length; i++) {
    var $nav = this.$navigation[i];
    $nav.$menuButton = $nav.querySelector('.js-x-header-toggle')
    
    $nav.$menu = $nav.$menuButton && $nav.querySelector(
      '#' + $nav.$menuButton.getAttribute('aria-controls')
      )
    if (!$nav.$menuButton || !$nav.$menu) {
      return
    }
    $nav.$menuOpenClass = $nav.$menu && $nav.$menu.dataset.openClass;
    $nav.$menuButtonOpenClass = $nav.$menuButton && $nav.$menuButton.dataset.openClass;
    $nav.$menuButtonOpenLabel = $nav.$menuButton && $nav.$menuButton.dataset.labelForShow;
    $nav.$menuButtonCloseLabel = $nav.$menuButton && $nav.$menuButton.dataset.labelForHide;
    $nav.$menuButtonOpenText = $nav.$menuButton && $nav.$menuButton.dataset.textForShow;
    $nav.$menuButtonCloseText = $nav.$menuButton && $nav.$menuButton.dataset.textForHide;
    $nav.isOpen = false;

    $nav.$menuButton.addEventListener('click', this.handleMenuButtonClick.bind($nav))
  }
}


/**
* Handle menu button click
*
* When the menu button is clicked, change the visibility of the menu and then
* sync the accessibility state and menu button state
*/
CrossServiceHeader.prototype.handleMenuButtonClick = function () {
  this.isOpen = !this.isOpen
  this.$menuOpenClass && this.$menu.classList.toggle(this.$menuOpenClass, this.isOpen)
  this.$menuButtonOpenClass && this.$menuButton.classList.toggle(this.$menuButtonOpenClass, this.isOpen)
  this.$menuButton.setAttribute('aria-expanded', this.isOpen)
  if (this.$menuButtonCloseLabel && this.$menuButtonOpenLabel) {
    this.$menuButton.setAttribute('aria-label', (this.isOpen ? this.$menuButtonCloseLabel : this.$menuButtonOpenLabel))
  }
  if (this.$menuButtonCloseText && this.$menuButtonOpenText) {
    this.$menuButton.innerHTML = this.isOpen ? this.$menuButtonCloseText : this.$menuButtonOpenText
  }
}
