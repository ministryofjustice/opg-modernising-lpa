describe('Dashboard', () => {
    context('with incomplete LPA', () => {
        beforeEach(() => {
            cy.visit('/fixtures/dashboard?asDonor=1&redirect=/dashboard');
        });

        it('shows my lasting power of attorney', () => {
            cy.checkA11yApp();

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'Drafting');
            cy.contains('a', 'Go to task list').click();

            cy.url().should('contain', '/task-list');
        });

        it('can create another LPA', () => {
            cy.contains('button', 'Start now').click();

            cy.url().should('contain', '/make-a-new-lpa');
            cy.checkA11yApp();
        });
    })

    context('when payment task started', () => {
        it('LPAs have a track progress button', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=checkAndSendToYourCertificateProvider');

            cy.contains('a', 'Pay for the LPA').click()
            cy.contains('a', 'Continue').click()
            cy.get('input[name="yes-no"]').check('yes', { force: true });
            cy.contains('button', 'Save and continue').click()

            cy.contains('a', 'Manage your LPAs').click()

            cy.get('button').should('not.contain', 'Continue');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list');
            cy.contains('a', 'Delete LPA');
            cy.contains('a', 'Check LPA progress').click();

            cy.url().should('contain', '/progress');
        });
    });

    context('when paid', () => {
        it('LPAs have a track progress button', () => {
            cy.visit('/fixtures?redirect=&progress=payForTheLpa');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list');
            cy.contains('a', 'Delete LPA');
            cy.contains('a', 'Check LPA progress').click();

            cy.url().should('contain', '/progress');
        });
    });

    context('with submitted LPA', () => {
        it('completed LPAs have a track progress button', () => {
            Cypress.on('uncaught:exception', () => {
                // TODO: remove this if this test works without, it is a problem
                // in the moj-frontend package
                return false
            })

            cy.visit('/fixtures?redirect=&progress=signTheLpa');

            cy.get('button').should('not.contain', 'Continue');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.get('.app-dashboard-row .moj-button-menu__item').should('have.length', 3);
            cy.contains('a', 'Go to task list');
            cy.contains('button', 'Actions').click();
            cy.contains('a', 'View LPA');
            cy.contains('a', 'Revoke LPA');
            cy.contains('a', 'Check LPA progress').click();

            cy.url().should('contain', '/progress');
        });
    });

    context('with statutory waiting period LPA', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=statutoryWaitingPeriod');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'Waiting period');
            cy.get('.app-dashboard-row a').should('have.length', 3);
            cy.contains('a', 'View LPA');
            cy.contains('a', 'Check LPA progress');
            cy.contains('a', 'Revoke LPA');
        });
    });

    context('with withdrawn LPA', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=withdrawn');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'Revoked');
            cy.contains('.app-dashboard-card a').should('not.exist');
            cy.get('.app-dashboard-row a').should('have.length', 1);
            cy.contains('a', 'View LPA');
        });
    });

    context('with registered LPA', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=registered');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'Registered');
            cy.contains('a', 'Continue to Use a lasting power of attorney');
        });
    });

    context('with various roles', () => {
        it('shows all of my LPAs', () => {
            cy.visit('/fixtures/dashboard?asDonor=1&asAttorney=1&asCertificateProvider=1&redirect=/dashboard');

            cy.contains('My LPAs');
            cy.contains('I’m an attorney');
            cy.contains('I’m a certificate provider');
        });
    })

    context('with an LPA that cannot be registered', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=certificateProviderOptedOut');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'Cannot register');
            cy.contains('.app-dashboard-card a').should('not.exist');
            cy.get('.app-dashboard-row a').should('have.length', 2);
            cy.contains('a', 'View LPA');
            cy.contains('a', 'Check LPA');
        });
    });
});
