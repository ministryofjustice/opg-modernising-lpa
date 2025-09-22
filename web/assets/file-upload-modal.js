import {ModalHelper} from "./modal-helper.js";

export class FileUploadModal {
    constructor(trigger, dialog) {
        this.dialog = dialog
        this.modal = new ModalHelper(trigger, dialog)
    }

    init() {
        this.cancelUploadButton = document.getElementById('cancel-upload-button')
        this.fileCount = document.getElementById('file-count')
        this.eventSource = null

        if (this.cancelUploadButton && this.modal.valid()) {
            this.registerListeners()

            if (this.dialog.dataset.startScan === "1") {
                this.modal.toggleVisibility()
                this.openConnection()
            }
        }
    }

    registerListeners() {
        this.cancelUploadButton.addEventListener('click', () => {
            document.getElementById('cancel-upload-form').submit()
        })

        this.dialog.addEventListener("close", () => {
            this.dialog.removeEventListener("keydown", this.modal.trapFocus.bind(this));
        });

        this.dialog.addEventListener("keydown", this.modal.trapFocus.bind(this));
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
