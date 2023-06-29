export class DataLossWarning {
    init() {
        this.returnToTaskListButton = document.getElementById('return-to-tasklist-btn')
        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')
        // so we can reference the same func when removing event
        this.handleTrapFocus = this.handleTrapFocus.bind(this)

        if (this.dialog && this.dialogOverlay && this.returnToTaskListButton) {
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
            this.dialogOverlay.classList.toggle('govuk-!-display-none')

            if (this.dialogVisible()) {
                this.dialog.addEventListener('keydown', this.handleTrapFocus)
                document.getElementById('back-to-page-btn').focus()
            } else {
                this.dialog.removeEventListener('keydown', this.handleTrapFocus)
                this.returnToTaskListButton.focus()
            }
        }
    }

    dialogVisible() {
        return !this.dialog.classList.contains('govuk-!-display-none') && !this.dialogOverlay.classList.contains('govuk-!-display-none')
    }

    registerListeners() {
        document.getElementById('return-to-tasklist-btn').addEventListener('click', (e) => {
            if (this.changesMade()) {
                e.preventDefault()
            }
        })
        document.getElementById('return-to-tasklist-btn').addEventListener('click', this.toggleDialogVisibility.bind(this))
        document.getElementById('back-to-page-btn').addEventListener('click', this.toggleDialogVisibility.bind(this))
        document.getElementById('return-to-task-list-dialog-btn').addEventListener('click', this.toggleDialogVisibility.bind(this))
    }

    handleTrapFocus(e) {
        const firstFocusableEl = document.getElementById('dialog-title')
        const lastFocusableEl = document.getElementById('return-to-task-list-dialog-btn')
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
