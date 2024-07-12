export class DataLossWarning {
  constructor(trigger, dialog) {
    this.trigger = trigger
    this.dialog = dialog
  }

  init() {
    this.dialogOverlay = document.getElementById('dialog-overlay')
    // so we can reference the same func when removing event
    this.handleTrapFocus = this.handleTrapFocus.bind(this)

    if (this.dialog && this.dialogOverlay && this.trigger) {
      this.submitButton = document.querySelector('button[type=submit]')
      this.backToPage = this.dialog.querySelector('button')
      this.continueWithoutSaving = this.dialog.querySelector('a')
      this.formValuesOnPageLoad = this.getEncodedStringifiedFormValues()
      this.formValuesPriorToPageLoad = this.getFormValuesFromCookie()
      this.registerListeners()
    }
  }

  changesMade() {
    return this.formValuesOnPageLoad !== this.getEncodedStringifiedFormValues() ||
      // to account for page reload on validation error
      this.formValuesPriorToPageLoad === this.getEncodedStringifiedFormValues()
  }

  formEmpty() {
    const encodedEmptyFormValues = encodeURIComponent(JSON.stringify([]))

    return this.getEncodedStringifiedFormValues() === encodedEmptyFormValues
  }

  getEncodedStringifiedFormValues() {
    const formValues = new FormData(document.querySelector("form:not([action])"))
    formValues.delete('csrf')

    const sanitisedValues = [...formValues.values()].filter(subArray => subArray.length > 0)

    return encodeURIComponent(JSON.stringify(sanitisedValues))
  }

  toggleDialogVisibility() {
    if (this.changesMade() && !this.formEmpty()) {
      this.dialog.classList.toggle('govuk-!-display-none')
      this.dialogOverlay.classList.toggle('govuk-!-display-none')

      if (this.dialogVisible()) {
        this.dialog.addEventListener('keydown', this.handleTrapFocus)
        this.backToPage.focus()
      } else {
        this.dialog.removeEventListener('keydown', this.handleTrapFocus)
        this.trigger.focus()
      }
    }
  }

  dialogVisible() {
    return !this.dialog.classList.contains('govuk-!-display-none') && !this.dialogOverlay.classList.contains('govuk-!-display-none')
  }

  registerListeners() {
    this.trigger.addEventListener('click', (e) => {
      if (this.changesMade() && !this.formEmpty()) {
        e.preventDefault()
      }
    })
    this.submitButton.addEventListener('click', this.addFormValuesToCookie.bind(this))
    this.trigger.addEventListener('click', this.toggleDialogVisibility.bind(this))
    this.backToPage.addEventListener('click', this.toggleDialogVisibility.bind(this))
    this.continueWithoutSaving.addEventListener('click', this.toggleDialogVisibility.bind(this))
  }

  handleTrapFocus(e) {
    const focusable = this.dialog.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])')
    const firstFocusableEl = focusable[0]
    const lastFocusableEl = focusable[focusable.length - 1]

    const KEY_CODE_TAB = 9
    const KEY_CODE_ESC = 27

    const tabPressed = (e.key === 'Tab' || e.keyCode === KEY_CODE_TAB)
    const escPressed = (e.key === 'Esc' || e.keyCode === KEY_CODE_ESC)

    if (tabPressed) {
      if (e.shiftKey) { /* shift + tab */
        if (document.activeElement === firstFocusableEl) {
          lastFocusableEl.focus()
          e.preventDefault()
        }
      } else /* tab */ {
        if (document.activeElement === lastFocusableEl) {
          firstFocusableEl.focus()
          e.preventDefault()
        }
      }
    }

    if (escPressed) {
      this.toggleDialogVisibility()
    }
  }

  addFormValuesToCookie() {
    // so the cookie isn't available for longer than required
    const tenSecondsFutureDate = new Date()
    tenSecondsFutureDate.setSeconds(tenSecondsFutureDate.getSeconds() + 10)

    document.cookie = `formValues=${this.getEncodedStringifiedFormValues()}; expires=${tenSecondsFutureDate.toUTCString()}; SameSite=Lax; Secure`
  }

  getFormValuesFromCookie() {
    return document.cookie.split("; ").find((row) => row.startsWith("formValues="))?.split("=")[1]
  }
}
