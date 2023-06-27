import { DataLossWarning } from './data-loss-warning'

const validBody = `
<div id="dialog-overlay" class="govuk-!-display-none" tabindex="-1"></div>
<div id="dialog"
     class="govuk-!-display-none govuk-!-padding-left-4 govuk-!-padding-top-2"
     role="dialog"
     aria-labelledby="dialog-title"
     aria-describedby="dialog-description"
     aria-modal="true">
    <div id="dialog-focus" tabindex="0"></div>
    <h2 id="dialog-title" class="govuk-heading-l">You have unsaved changes</h2>
    <p id="dialog-description" class="govuk-body">To save, go back to the page and select <span class="govuk-body govuk-!-font-weight-bold">Save and continue</span>.</p>

    <div class="govuk-button-group" data-module="app-save-or-return">
        <button type="button" id='back-to-page' class="govuk-button" data-module="govuk-button" aria-label="Back to page">Back to page</button>
        <a href="/task-list" id='return-to-task-list-dialog' class="govuk-button govuk-button--secondary">Continue without saving</a>
    </div>
</div>

<form>
    <input type="text">
    <textarea></textarea>
    <div class="govuk-button-group" data-module="app-save-or-return">
        <button id='submit-btn' class="govuk-button" data-module="govuk-button">Save and continue</button>
        <a href="/task-list" id='return-to-task-list-form' class="govuk-button govuk-button--secondary">Return to task list</a>
    </div>
</form>
`

describe('component validation', () => {
    describe('save or return', () => {
        describe('valid when', () => {
            it.each([
                {
                    name: 'anchor and button',
                    body: `<form>
                        <div class="govuk-button-group" data-module="app-save-or-return">
                            <button class="govuk-button" data-module="govuk-button">Save and continue</button>
                            <a href="/" class="govuk-button govuk-button--secondary">Return to task list</a>
                        </div>
                    </form>`,
                },
                {
                    name: 'two anchors',
                    body: `<form>
                        <div class="govuk-button-group" data-module="app-save-or-return">
                            <a href="/" class="govuk-button govuk-button--secondary">Save and continue</a>
                            <a href="/" class="govuk-button govuk-button--secondary">Return to task list</a>
                        </div>
                    </form>`,
                },
                {
                    name: 'two buttons',
                    body: `<form>
                        <div class="govuk-button-group" data-module="app-save-or-return">
                            <button class="govuk-button" data-module="govuk-button">Save and continue</button>
                            <button class="govuk-button" data-module="govuk-button">Return to task list</button>
                        </div>
                    </form>`,
                },
            ])('$name', ({name, body}) => {
                document.body.innerHTML = body
                const sut = new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`))

                expect(sut.saveOrReturnComponentValid()).toEqual(true)
            })
        })

        describe('invalid when', () => {
            it.each([
                {
                    name: 'one child element',
                    body: `<form>
                        <div class="govuk-button-group" data-module="app-save-or-return">
                            <button class="govuk-button" data-module="govuk-button">Save and continue</button>
                        </div>
                    </form>`,
                },
                {
                    name: 'wrong data-module',
                    body: `<form>
                        <div class="govuk-button-group" data-module="not-app-save-or-return">
                            <button class="govuk-button" data-module="govuk-button">Save and continue</button>
                            <a href="/" class="govuk-button govuk-button--secondary">Return to task list</a>
                        </div>
                    </form>`,
                },
                {
                    name: 'wrong child element',
                    body: `<form>
                        <div class="govuk-button-group" data-module="app-save-or-return">
                            <div class="govuk-button" data-module="govuk-button">Save and continue</div>
                            <p class="govuk-button govuk-button--secondary">Return to task list</p>
                        </div>
                    </form>`,
                },
            ])('$name', ({name, body}) => {
                document.body.innerHTML = body
                const sut = new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`))

                expect(sut.saveOrReturnComponentValid()).toEqual(false)
            })
        })
    })

    describe('dialog', () => {
        describe('valid when', () => {
            it('expected divs, anchors and buttons are present', () => {
                document.body.innerHTML = `<div id="dialog-overlay"></div>
                <div id="dialog">
                    <div class="govuk-button-group">
                        <button class="govuk-button" data-module="govuk-button" aria-label="Close Navigation">Back to page</button>
                        <a href="/task-list" id='return-to-task-list-popup' class="govuk-button govuk-button--secondary">Continue without saving</a>
                    </div>
                </div>`
                const sut = new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`))
                sut.init()

                expect(sut.dialogComponentValid()).toEqual(true)
            })
        })

        describe('invalid when', () => {
            it.each([
                {
                    name: 'overlay div missing',
                    body: `<div id="not-dialog-overlay"></div>
                    <div id="dialog">
                        <div class="govuk-button-group">
                            <button class="govuk-button" data-module="govuk-button" aria-label="Close Navigation">Back to page</button>
                            <a href="/task-list" id='return-to-task-list-popup' class="govuk-button govuk-button--secondary">Continue without saving</a>
                        </div>
                    </div>`,
                },
                {
                    name: 'dialog div missing',
                    body: `<div id="dialog-overlay"></div>
                    <div id="not-dialog">
                        <div class="govuk-button-group">
                            <button class="govuk-button" data-module="govuk-button" aria-label="Close Navigation">Back to page</button>
                            <a href="/task-list" id='return-to-task-list-popup' class="govuk-button govuk-button--secondary">Continue without saving</a>
                        </div>
                    </div>`,
                },
                {
                    name: 'wrong number of anchors/buttons',
                    body: `<div id="dialog-overlay"></div>
                    <div id="dialog">
                        <div class="govuk-button-group">
                            <button class="govuk-button" data-module="govuk-button" aria-label="Close Navigation">Back to page</button>
                        </div>
                    </div>`,
                },
            ])('$name', ({name, body}) => {
                document.body.innerHTML = body
                const sut = new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`))

                expect(sut.dialogComponentValid()).toEqual(false)
            })
        })

        describe('toggling visibility', () => {
            it('toggles required classes', () => {
                document.body.innerHTML = validBody
                const sut = new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`))
                sut.init()

                sut.changesMade = jest.fn().mockReturnValue(true)
                sut.dialogVisible = jest.fn().mockReturnValue(true)

                const dialog = document.getElementById('dialog')
                const dialogOverlay = document.getElementById('dialog-overlay')

                console.log(dialog.innerHTML)
                expect(dialog.classList.contains('govuk-!-display-none')).toBeTruthy()
                expect(dialogOverlay.classList.contains('govuk-!-display-none')).toBeTruthy()
                expect(dialog.classList.contains('dialog')).toBeFalsy()
                expect(dialogOverlay.classList.contains('dialog-overlay')).toBeFalsy()

                sut.toggleDialogVisibility()

                expect(dialog.classList.contains('govuk-!-display-none')).toBeFalsy()
                expect(dialogOverlay.classList.contains('govuk-!-display-none')).toBeFalsy()
                expect(dialog.classList.contains('dialog')).toBeTruthy()
                expect(dialogOverlay.classList.contains('dialog-overlay')).toBeTruthy()

                sut.dialogVisible = jest.fn().mockReturnValue(false)
                sut.toggleDialogVisibility()

                expect(dialog.classList.contains('govuk-!-display-none')).toBeTruthy()
                expect(dialogOverlay.classList.contains('govuk-!-display-none')).toBeTruthy()
                expect(dialog.classList.contains('dialog')).toBeFalsy()
                expect(dialogOverlay.classList.contains('dialog-overlay')).toBeFalsy()
            })
        })
    })
})

// limitations with FormData API + jsdom + jest means testing toggling visibility falls to cypress
