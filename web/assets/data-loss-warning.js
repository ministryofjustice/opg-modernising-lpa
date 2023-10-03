export class DataLossWarning {
    init() {
        this.returnToTaskListButton = document.getElementById('return-to-tasklist-btn')
        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')
        // so we can reference the same func when removing event
        this.handleTrapFocus = this.handleTrapFocus.bind(this)

        if (this.dialog && this.dialogOverlay && this.returnToTaskListButton) {
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
                document.getElementById('back-to-page-dialog-btn').focus()
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
            if (this.changesMade() && !this.formEmpty()) {
                e.preventDefault()
            }
        })
        document.getElementById('save-and-continue-btn').addEventListener('click', this.addFormValuesToCookie.bind(this))
        document.getElementById('return-to-tasklist-btn').addEventListener('click', this.toggleDialogVisibility.bind(this))
        document.getElementById('back-to-page-dialog-btn').addEventListener('click', this.toggleDialogVisibility.bind(this))
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
