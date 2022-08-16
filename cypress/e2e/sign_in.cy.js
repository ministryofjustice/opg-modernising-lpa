const _ = Cypress._

describe('Sign in using GOV UK Sign In service', () => {
    Cypress.Commands.add('loginBySingleSignOn', (overrides = {}) => {
        Cypress.log({
            name: 'loginBySingleSignOn',
        })

        const options = {
            method: 'GET',
            url: 'http://localhost:7011/authorize',
            qs: {
                redirect_uri: 'http://localhost:5050/set_token',
                client_id: 'client-credentials-mock-client',
                state: 'state-value',
                nonce: 'nonce-value',
                scope: 'scope-value'
            },
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

    context('with an existing GOV UK account', () => {
        context('Use redirectTo and a session cookie to login', function () {
            // it('is 403 unauthorized without a session cookie', function () {
            //     // smoke test just to show that without logging in we cannot
            //     // visit the landing page
            //     cy.visit('/home')
            //     cy.get('h3').should(
            //         'contain',
            //         'You are not logged in and cannot access this page'
            //     )
            //
            //     cy.url().should('include', 'unauthorized')
            // })

            it('can authenticate with cy.request', function () {
                // before we start, there should be no session cookie
                cy.getCookie('sign-in-token').should('not.exist')

                // this automatically gets + sets cookies on the browser
                // and follows all of the redirects that ultimately get
                // us to /dashboard.html
                cy.loginBySingleSignOn().then((resp) => {
                    expect(resp.status).to.eq(200)
                    expect(resp.body).to.include('Welcome gideon.felix@example.org')
                })

                // the redirected page hits the OIDC, and the OIDC middleware
                // parses the authentication token and returns the dashboard view
                // with our cookie 'cypress-session-cookie' set
                cy.getCookie('sign-in-token').should('exist')

                // you don't need to do this next part but
                // just to prove we can also visit the page in our app
                cy.visit('/home')

                cy.get('h1').should('contain', 'Welcome gideon.felix@example.org')
            })
        })
    })

})
