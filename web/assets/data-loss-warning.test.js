import returnToTaskListAndSaveButtonsVisible from './data-loss-warning'

describe('data loss warning', () => {
    it('knows when a return to task list and save and continue button are part of a form', () => {
        document.body.innerHTML = `
<form>
    <div class="govuk-button-group" id="save-or-return">
        <button class="govuk-button" data-module="govuk-button">Some button text</button>
        <a href="/" class="govuk-button govuk-button--secondary">More button text</a>
    </div>
</form>`

        expect(returnToTaskListAndSaveButtonsVisible()).toBeTruthy()
    })


})
