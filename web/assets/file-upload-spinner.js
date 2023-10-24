export class FileUploadSpinner {
    init() {
        this.continueButton = document.getElementById('continue-or-pay')
        this.returnToTaskListButton = document.getElementById('return-to-tasklist-btn')
        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')
        // so we can reference the same func when removing event
        this.handleTrapFocus = this.handleTrapFocus.bind(this)

        this.registerListeners()
    }

    registerListeners() {
        this.continueButton.addEventListener('click', (e) => {e.preventDefault()})
        this.continueButton.addEventListener('click', this.toggleDialogVisibility.bind(this))

        document.getElementById('cancel-upload-button').addEventListener('click', this.toggleDialogVisibility.bind(this))
    }

    dialogVisible() {
        return !this.dialog.classList.contains('govuk-!-display-none') && !this.dialogOverlay.classList.contains('govuk-!-display-none')
    }

    toggleDialogVisibility() {
        this.dialog.classList.toggle('govuk-!-display-none')
        this.dialogOverlay.classList.toggle('govuk-!-display-none')

        if (this.dialogVisible()) {
            this.dialog.addEventListener('keydown', this.handleTrapFocus)
            this.continueButton.focus()
        } else {
            this.dialog.removeEventListener('keydown', this.handleTrapFocus)
            this.returnToTaskListButton.focus()
        }
    }

    handleTrapFocus(e) {
        const firstFocusableEl = document.getElementById('dialog-title')
        const lastFocusableEl = document.getElementById('cancel-upload-button')
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
}
