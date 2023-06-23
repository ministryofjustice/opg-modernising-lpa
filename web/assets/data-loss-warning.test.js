import { DataLossWarning } from './data-loss-warning'

const validBody = `
<div id="dialog-overlay" class="govuk-visually-hidden" tabindex="-1"></div>
<div id="dialog"
        class="govuk-visually-hidden govuk-!-padding-left-4 govuk-!-padding-top-2"
        role="dialog"
        aria-labelledby="dialog-title"
        aria-describedby="dialog-description"
        aria-modal="true"
        style="border: 10px solid black;">
    <div tabindex="0" class="dialog-focus"></div>
    <h2 id="dialog-title" class="govuk-heading-l">You have unsaved changes</h2>
    <p id="dialog-description" class="govuk-body">To save, go back to the page and select <span class="govuk-body govuk-!-font-weight-bold">Save and continue</span>.</p>

    <div class="govuk-button-group">
        <button type="button" id='back-to-page-btn' class="govuk-button" data-module="govuk-button" aria-label="Close Navigation">Back to page</button>
        <a href="/task-list" id='return-to-task-list-popup' class="govuk-button govuk-button--secondary">Continue without saving</a>
    </div>
</div>

<form>
    <input id="input" type="text" value="hello">
    <textarea id="textarea"></textarea>
    <select id="select"><option value="1">Option 1</option></select>
    <input id="radio" type="radio" value="2">
    <input id="checkbox" type="checkbox" value="3">

    <div class="govuk-button-group" data-module="app-save-or-return">
        <button id='submit-btn' class="govuk-button" data-module="govuk-button">Save and continue</button>
        <a href="/task-list" id='return-to-task-list-form' class="govuk-button govuk-button--secondary">Return to task list</a>
    </div>
</form>`

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
    })

})

describe('toggling popup visiblity', () => {
    it.each([
        {
            elementId: 'input',
            eventType: 'keyup',
        },
        {
            elementId: 'textarea',
            eventType: 'keyup',
        },
        {
            elementId: 'checkbox',
            eventType: 'change',
        },
        {
            elementId: 'radio',
            eventType: 'change',
        },
        {
            elementId: 'select',
            eventType: 'change',
        },
    ])('shown if changes have been made to $elementId', ({elementId, eventType}) => {
        document.body.innerHTML = validBody
        const fd = new FormData(document.querySelector('form'))
        fd.append('s', 'v')
        console.log(fd.values())


        new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`)).init()

        const element = document.getElementById(elementId)
        element.value = "hi"
        element.dispatchEvent(new Event('input', { bubbles: true }))

        document.getElementById('return-to-task-list-form').click()

        const popUpOverlay = document.getElementById('dialog-overlay')
        const popUp = document.getElementById('dialog')

        expect(popUpOverlay.classList.contains('govuk-visually-hidden')).toEqual(false)
        expect(popUp.classList.contains('govuk-visually-hidden')).toEqual(false)
    })

    it('not shown if changes have not been made', () => {
        document.body.innerHTML = validBody

        new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`)).registerListeners()

        document.getElementById('return-to-task-list-form').click()

        const popUpOverlay = document.getElementById('dialog-overlay')
        const popUp = document.getElementById('dialog')

        expect(popUpOverlay.classList.contains('govuk-visually-hidden')).toEqual(true)
        expect(popUp.classList.contains('govuk-visually-hidden')).toEqual(true)
    })
})

describe('interacting with pop up', () => {
    it('clicking back to page popup button hides overlay', () => {
        document.body.innerHTML = validBody

        const sut = new DataLossWarning(document.querySelector(`[data-module="app-save-or-return"]`))

        document.querySelector('input').dispatchEvent(new Event('change', { bubbles: true }))
        document.getElementById('return-to-task-list-form').click()
        document.getElementById('back-to-page-btn').click()

        const popUpOverlay = document.getElementById('dialog-overlay')
        const popUp = document.getElementById('dialog')

        expect(popUpOverlay.classList.contains('govuk-visually-hidden')).toEqual(true)
        expect(popUp.classList.contains('govuk-visually-hidden')).toEqual(true)
    })
})
