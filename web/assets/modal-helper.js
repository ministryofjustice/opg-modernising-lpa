export class ModalHelper {
    constructor (trigger, dialog) {
        this.trigger = trigger
        this.dialog = dialog
    }

    valid() {
        return this.dialog && this.trigger
    }

    trapFocus(e) {
        if (e.key !== "Tab") return;

        const focusable = this.dialog.querySelectorAll('button, [href]')
        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey && document.activeElement === first) {
            e.preventDefault();
            last.focus();
        }
        else if (!e.shiftKey && document.activeElement === last) {
            e.preventDefault();
            first.focus();
        }
    }

    toggleVisibility() {
        if (this.dialog.open) {
            this.dialog.close()
            this.trigger.focus()
        } else {
            this.dialog.showModal()
        }
    }
}
