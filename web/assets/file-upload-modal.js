export class FileUploadModal {
    constructor(dialog) {
        this.dialog = dialog
    }

    init() {
        this.cancelUploadButton = document.getElementById('cancel-upload-button')
        this.fileCount = document.getElementById('file-count')
        this.eventSource = null

        if (this.cancelUploadButton && this.dialog) {
            this.registerListeners()

            if (this.dialog.dataset.startScan === "1") {
                this.toggleVisibility()
                this.openConnection()
            }
        }
    }

    toggleVisibility() {
        if (this.dialog.open) {
            this.dialog.close()
        } else {
            this.dialog.showModal()
        }
    }

    registerListeners() {
        this.cancelUploadButton.addEventListener('click', () => {
            document.getElementById('cancel-upload-form').submit()
        })
    }

    openConnection() {
        this.eventSource = new EventSource(document.querySelector("[data-sse-url]").dataset.sseUrl);

        this.eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data)

            if (data.finishedScanning === true) {
                document.getElementById('scan-results-form').submit()
            }

            if (data.closeConnection === "1") {
                this.eventSource.close()
                document.getElementById('close-connection-form').submit()
            }

            let parts = this.fileCount.innerHTML.split(' ')
            parts[0] = data.scannedCount
            this.fileCount.innerHTML = parts.join(' ')
        };
    }
}
