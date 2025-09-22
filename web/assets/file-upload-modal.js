export class FileUploadModal {
    constructor(modal) {
        this.modal = modal
    }

    init() {
        this.cancelUploadButton = document.getElementById('cancel-upload-button')
        this.fileCount = document.getElementById('file-count')
        this.eventSource = null

        if (this.cancelUploadButton && this.modal) {
            this.registerListeners()

            if (this.modal.dataset.startScan === "1") {
                this.toggleVisibility()
                this.openConnection()
            }
        }
    }

    toggleVisibility() {
        if (this.modal.open) {
            this.modal.close()
        } else {
            this.modal.showModal()
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
