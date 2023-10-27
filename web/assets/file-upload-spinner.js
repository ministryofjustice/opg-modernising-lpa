export class FileUploadSpinner {
    init() {
        // so we can reference the same func when removing event
        this.handleTrapFocus = this.handleTrapFocus.bind(this)

        this.continueButton = document.getElementById('continue-or-pay')
        this.cancelUploadButton = document.getElementById('cancel-upload-button')

        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')
        this.dialogTitle = document.getElementById('dialog-title')
        this.dialogFileCount = document.getElementById('file-count')

        this.eventSource = null

        if (this.continueButton && this.cancelUploadButton) {
            this.registerListeners()
        }
    }

    registerListeners() {
        this.continueButton.addEventListener('click', (e) => {e.preventDefault()})
        this.continueButton.addEventListener('click', this.toggleDialogVisibility.bind(this))
        this.continueButton.addEventListener('click', this.openConnection.bind(this))

        this.cancelUploadButton.addEventListener('click', this.toggleDialogVisibility.bind(this))
        this.cancelUploadButton.addEventListener('click', this.closeConnection.bind(this))
    }

    toggleDialogVisibility() {
        this.dialog.classList.toggle('govuk-!-display-none')
        this.dialogOverlay.classList.toggle('govuk-!-display-none')

        if (this.dialogVisible()) {
            this.dialog.addEventListener('keydown', this.handleTrapFocus)
            this.dialogTitle.focus()
        } else {
            this.dialog.removeEventListener('keydown', this.handleTrapFocus)
        }
    }

    dialogVisible() {
        return !this.dialog.classList.contains('govuk-!-display-none') && !this.dialogOverlay.classList.contains('govuk-!-display-none')
    }

    openConnection() {
        this.eventSource = new EventSource(document.querySelector("[data-sse-url]").dataset.sseUrl);

        this.eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data)

            if (data.scannedTotal === data.fileTotal) {
                document.getElementById('pay-form').submit()
            }

            let parts = this.dialogFileCount.innerHTML.split(' ')
            parts[0] = data.scannedTotal
            this.dialogFileCount.innerHTML = parts.join(' ')
        };
    }

    closeConnection() {
        this.eventSource.close()
    }

    handleTrapFocus(e) {
        const firstFocusableEl = this.dialogTitle
        const lastFocusableEl = this.cancelUploadButton
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
            this.closeConnection()
            this.toggleDialogVisibility()
        }
    }
}
