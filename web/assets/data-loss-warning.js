export default class DataLossWarning {

    constructor(saveOrReturnComponent) {
        this.saveOrReturnComponent = saveOrReturnComponent
        this.changesMade = false
        this.registerListeners()
    }

    setChangesMade() {
        this.changesMade = true
    }

    togglePopupVisibility() {
        if (this.changesMade) {
            document.getElementById('dialog-overlay').classList.toggle('hide')
            document.getElementById('dialog').classList.toggle('hide')
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
                if (['INPUT', 'TEXTAREA'].includes(element.tagName)) {
                    element.addEventListener('change', this.setChangesMade.bind(this))
                }

                if (element.tagName === 'A') {
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
