import hasReturnToTaskListAndSaveButtons from './data-loss-warning'

describe('data loss warning', () => {
    it.each([
        {
            body: `<form>
                <div class="govuk-button-group" data-module="app-save-or-return">
                    <button class="govuk-button" data-module="govuk-button">Some button text</button>
                    <a href="/" class="govuk-button govuk-button--secondary">More button text</a>
                </div>
            </form>`,
            expected: true,
        },
        {
            body: `<form>
                <div class="govuk-button-group" data-module="app-save-or-return">
                    <button class="govuk-button" data-module="govuk-button">Some button text</button>
                </div>
            </form>`,
            expected: false,
        },
        {
            body: `<form>
                <div class="govuk-button-group" data-module="not-app-save-or-return">
                    <button class="govuk-button" data-module="govuk-button">Some button text</button>
                    <a href="/" class="govuk-button govuk-button--secondary">More button text</a>
                </div>
            </form>`,
            expected: false,
        },
        {
            body: `<form>
                <div class="govuk-button-group" data-module="app-save-or-return">
                    <div class="govuk-button" data-module="govuk-button">Some button text</div>
                    <p class="govuk-button govuk-button--secondary">More button text</p>
                </div>
            </form>`,
            expected: false,
        },
    ])('knows when a save or return button group has two buttons with body %s', ({body, expected}) => {
        document.body.innerHTML = body
        expect(hasReturnToTaskListAndSaveButtons()).toBe(expected)
    })

})
