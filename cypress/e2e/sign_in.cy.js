const _ = Cypress._

describe('Sign in using GOV UK Sign In service', () => {
    Cypress.Commands.add('loginBySingleSignOn', (overrides = {}) => {
        Cypress.log({
            name: 'loginBySingleSignOn',
        })

        const options = {
            method: 'GET',
            url: '/login',
            log: true
        }

        // allow us to override defaults with passed in overrides
        _.extend(options, overrides)

        cy.request(options)
    })

    beforeEach(() => {
        cy.visit('/home')
        cy.injectAxe()
    })

    context('accessing home page without logging in', () => {
        it('does not show user email', () => {
            cy.visit('/home')
            cy.get('h1').should(
                'contain',
                'User not signed in'
            )
        })
    })

    context('with an existing GOV UK account', () => {
        context('accessing home page after logging in', () => {
            it('can authenticate with cy.request', () => {
                cy.loginBySingleSignOn().then((resp) => {
                    expect(resp.status).to.eq(200)
                    expect(resp.body).to.include('gideon.felix@example.org')
                })
            })
        })
    })

})
