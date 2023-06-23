export class DataLossWarning {
    constructor(saveOrReturnComponent) {
        this.saveOrReturnComponent = saveOrReturnComponent
        this.originalFormValues = ''
    }

    init() {
        this.originalFormValues = this.stringifyFormValues()
        this.registerListeners()
    }

    changesMade() {
        return this.originalFormValues !== this.stringifyFormValues()
    }

    stringifyFormValues() {
        return JSON.stringify(...new FormData(document.querySelector('form')).values())
    }

    togglePopupVisibility() {
        if (this.changesMade()) {
            document.getElementById('dialog-overlay').classList.toggle('govuk-visually-hidden')
            document.getElementById('dialog').classList.toggle('govuk-visually-hidden')
        }
    }

    saveOrReturnComponentValid() {
        if (!this.saveOrReturnComponent) {
            return false
        }

        return this.buttonsPresent(2, this.saveOrReturnComponent)
    }

    dialogComponentValid() {
        const overlay = document.getElementById('dialog-overlay')
        const dialog = document.getElementById('dialog')

        if (!overlay || !dialog) {
            return false
        }

        return this.buttonsPresent(2, dialog.querySelector(".govuk-button-group"))
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
        if (this.saveOrReturnComponentValid()) {
            for (let element of this.saveOrReturnComponent.children) {
                if (element.tagName === 'A') {
                    element.addEventListener('click', (e) => {
                        if (this.changesMade()) {
                            e.preventDefault()
                        }
                    })

                    element.addEventListener('click', this.togglePopupVisibility.bind(this))
                }
            }
        }

        if (this.dialogComponentValid()) {
            for (let element of document.getElementById('dialog').querySelector(".govuk-button-group").children) {
                if (element.tagName === 'BUTTON') {
                    element.addEventListener('click', this.togglePopupVisibility.bind(this))
                }
            }

        }
    }
}
