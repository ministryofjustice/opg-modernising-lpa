export class DataLossWarning {
    constructor(trigger, modal) {
        this.trigger = trigger
        this.modal = modal
    }

    init() {
        if (this.modal && this.trigger) {
            this.submitButton = document.querySelector('button[type=submit]')
            this.backToPage = this.modal.querySelector('button')
            this.continueWithoutSaving = this.modal.querySelector('a')
            this.formValuesOnPageLoad = this.getEncodedStringifiedFormValues()
            this.formValuesPriorToPageLoad = this.getFormValuesFromCookie()
            this.registerListeners()
        }
    }

    toggleVisibility() {
        if (this.modal.open) {
            this.modal.close()
        } else {
            this.modal.showModal()
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

    registerListeners() {
        this.trigger.addEventListener('click', (e) => {
            if (this.changesMade() && !this.formEmpty()) {
                e.preventDefault()
            }
        })

        this.trigger.addEventListener('click', this.toggleVisibility.bind(this))
        this.backToPage.addEventListener('click', this.toggleVisibility.bind(this))
        this.continueWithoutSaving.addEventListener('click', this.toggleVisibility.bind(this))
        this.submitButton.addEventListener('click', this.addFormValuesToCookie.bind(this))
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
