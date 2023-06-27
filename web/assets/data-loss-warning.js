export class DataLossWarning {
    constructor(saveOrReturnComponent) {
        this.saveOrReturnComponent = saveOrReturnComponent
        this.originalFormValues = ''
        this.dialog = null
        this.dialogOverlay = null
    }

    init() {
        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')
        // so we can reference the same func when removing event
        this.handleTrapFocus = this.handleTrapFocus.bind(this)

        if (this.dialogComponentValid() && this.saveOrReturnComponentValid()) {
            this.originalFormValues = this.stringifyFormValues()
            this.registerListeners()
        }
    }

    changesMade() {
        return this.originalFormValues !== this.stringifyFormValues()
    }

    stringifyFormValues() {
        return JSON.stringify([...new FormData(document.querySelector("form:not([action])")).values()])
    }

    toggleDialogVisibility() {
        if (this.changesMade()) {
            this.dialog.classList.toggle('govuk-!-display-none')
            this.dialog.classList.toggle('dialog')
            this.dialogOverlay.classList.toggle('govuk-!-display-none')
            this.dialogOverlay.classList.toggle('dialog-overlay')

            if (this.dialogVisible()) {
                this.dialog.addEventListener('keydown', this.handleTrapFocus)
                document.getElementById('dialog-focus').focus()
            } else {
                this.dialog.removeEventListener('keydown', this.handleTrapFocus)
                this.saveOrReturnComponent.querySelector('a').focus()
            }
        }
    }

    dialogVisible() {
        return !this.dialog.classList.contains('govuk-!-display-none') && !this.dialogOverlay.classList.contains('govuk-!-display-none')
    }

    saveOrReturnComponentValid() {
        if (!this.saveOrReturnComponent) {
            return false
        }

        return this.buttonsPresent(2, this.saveOrReturnComponent)
    }

    dialogComponentValid() {
        if (!this.dialog || !this.dialogOverlay) {
            return false
        }

        return this.buttonsPresent(2, this.dialog.querySelector(".govuk-button-group"))
    }

    buttonsPresent(requiredCount, parentElement) {
        let buttonCount = 0

        for (let element of parentElement.children) {
            if (['A', 'BUTTON'].includes(element.tagName)) {
                buttonCount++
            }
        }

        return buttonCount === requiredCount
    }

    registerListeners() {
        for (let element of this.saveOrReturnComponent.children) {
            if (element.tagName === 'A') {
                element.addEventListener('click', (e) => {
                    if (this.changesMade()) {
                        e.preventDefault()
                    }
                })

                element.addEventListener('click', this.toggleDialogVisibility.bind(this))
            }
        }

        for (let element of this.dialog.querySelector(".govuk-button-group").children) {
            if (element.tagName === 'BUTTON') {
                element.addEventListener('click', this.toggleDialogVisibility.bind(this))
            }
        }
    }

    handleTrapFocus(e) {
        const focusableEls = this.dialog.querySelectorAll('a[href], button')
        const firstFocusableEl = focusableEls[0]
        const lastFocusableEl = focusableEls[focusableEls.length - 1]
        const KEY_CODE_TAB = 9
        const KEY_CODE_ESC = 27

        const tabPressed = (e.key === 'Tab' || e.keyCode === KEY_CODE_TAB)
        const escPressed = (e.key === 'Esc' || e.keyCode === KEY_CODE_ESC)

        if (!tabPressed && !escPressed) {
            return;
        }

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
}
