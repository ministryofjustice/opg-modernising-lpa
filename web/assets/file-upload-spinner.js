export class FileUploadSpinner {
    init() {
        this.continueButton = document.getElementById('continue-or-pay')
        this.cancelUploadButton = document.getElementById('cancel-upload-button')

        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')

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
            console.log(event.data)
        };
    }

    closeConnection() {
        this.eventSource.close()
    }
}
