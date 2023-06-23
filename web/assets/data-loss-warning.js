export class DataLossWarning {
    constructor(saveOrReturnComponent) {
        this.saveOrReturnComponent = saveOrReturnComponent
        this.originalFormValues = ''
    }

    init() {
        this.dialog = document.getElementById('dialog')
        this.dialogOverlay = document.getElementById('dialog-overlay')

        if (this.dialogComponentValid() && this.saveOrReturnComponentValid()) {
            this.originalFormValues = this.stringifyFormValues()
            this.registerListeners()
        }
    }

    changesMade() {
        console.log(this.originalFormValues)
        console.log(this.stringifyFormValues())
        return this.originalFormValues !== this.stringifyFormValues()
    }

    stringifyFormValues() {
        return JSON.stringify([...new FormData(document.querySelector('form')).values()])
    }

    toggleDialogVisibility() {
        if (this.changesMade()) {
            this.dialog.classList.toggle('govuk-visually-hidden')
            this.dialog.classList.toggle('dialog')
            this.dialogOverlay.classList.toggle('govuk-visually-hidden')
            this.dialogOverlay.classList.toggle('dialog-overlay')
        }
    }

    saveOrReturnComponentValid() {
        if (!this.saveOrReturnComponent) {
            return false
        }

        return this.buttonsPresent(2, this.saveOrReturnComponent)
    }

    dialogComponentValid() {
        if (!this.dialog || !this.dialogOverlay) {
            return false
        }

        return this.buttonsPresent(2, this.dialog.querySelector(".govuk-button-group"))
    }

    buttonsPresent(requiredCount, parentElement) {
        let buttonCount = 0

        for (let element of parentElement.children) {
            if (['A', 'BUTTON'].includes(element.tagName)) {
                buttonCount++
            }
        }

        return buttonCount === requiredCount
    }

    registerListeners() {
        for (let element of this.saveOrReturnComponent.children) {
            if (element.tagName === 'A') {
                element.addEventListener('click', (e) => {
                    if (this.changesMade()) {
                        e.preventDefault()
                    }
                })

                element.addEventListener('click', this.toggleDialogVisibility.bind(this))
            }
        }

        for (let element of this.dialog.querySelector(".govuk-button-group").children) {
            if (element.tagName === 'BUTTON') {
                element.addEventListener('click', this.toggleDialogVisibility.bind(this))
            }
        }
    }
}
