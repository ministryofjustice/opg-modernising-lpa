export class FileUploadSpinner {
    init() {
        this.continueButton = document.getElementById('continue-or-pay')
        this.cancelUploadButton = document.getElementById('cancel-upload-button')

        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')
        this.dialogFileCount = document.getElementById('file-count')

        this.sseURL = document.querySelector("[data-sse-url]").dataset.sseUrl
        this.eventSource = null

        this.registerListeners()
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
    }

    openConnection() {
        this.eventSource = new EventSource(this.sseURL);

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
}
