import { oneLoginUrl, randomAccessCode } from "../../support/e2e.js";

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
            cy.contains('a', 'Make or add LPAs').click();

            cy.url().should('contain', '/make-or-add-an-lpa');
            cy.contains('button', 'Start').click();
            cy.checkA11yApp();

            cy.url().should('contain', '/make-a-new-lpa');
        });
    })

    context('when payment task started', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=checkAndSendToYourCertificateProvider');

            cy.contains('a', 'Pay for the LPA').click()
            cy.contains('a', 'Continue').click()
            cy.get('input[name="yes-no"]').check('yes', { force: true });
            cy.contains('button', 'Save and continue').click()

            cy.contains('a', 'Manage LPAs').click()

            cy.get('button').should('not.contain', 'Continue');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');

            cy.get('.app-dashboard-row a').should('have.length', 3);

            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list');
            cy.contains('a', 'Delete LPA');
            cy.contains('a', 'Check LPA progress').click();

            cy.url().should('contain', '/progress');
        });
    });

    context('when paid', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=payForTheLpa');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');

            cy.get('.app-dashboard-row a').should('have.length', 3);

            cy.contains('a', 'Go to task list');
            cy.contains('a', 'Delete LPA');
            cy.contains('a', 'Check LPA progress').click();

            cy.url().should('contain', '/progress');
        });
    });

    context('with signed LPA having pending task', () => {
        it('shows the correct options', () => {
            Cypress.on('uncaught:exception', () => {
                // TODO: remove this if this test works without, it is a problem
                // in the moj-frontend package
                return false
            })

            cy.visit('/fixtures?redirect=&progress=signTheLpa&idStatus=donor:insufficient-evidence');

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
            cy.contains('a', 'Go to task list');
        });
    });

    context('with submitted LPA', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=signTheLpa');

            cy.get('button').should('not.contain', 'Continue');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'View LPA');
            cy.contains('a', 'Revoke LPA');
            cy.contains('a', 'Check LPA progress').click();

            cy.url().should('contain', '/progress');
            cy.contains('a', 'Go to task list').should('not.exist');
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
            cy.get('.app-dashboard-row a').should('have.length', 2);
            cy.contains('a', 'View LPA');
            cy.contains('a', 'Check LPA progress');
        });
    });

    context('with registered LPA', () => {
        it('shows the correct options', () => {
            cy.visit('/fixtures?redirect=&progress=registered');

            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'Registered');
            cy.contains('a', 'Use a lasting power of attorney');
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

    context('with a donor access code', () => {
        it('can add a donor LPA', () => {
            const randomCode = randomAccessCode();
            cy.visit(`/fixtures/supporter?redirect=/dashboard&organisation=1&accessCode=${randomCode}`);

            cy.visit('/start')
            cy.contains('a', 'Start').click();
            cy.origin(oneLoginUrl(), () => {
                cy.contains('button', 'Continue').click();
            });
            cy.url().should('contain', '/make-or-add-an-lpa');

            cy.contains('a', 'Continue').click();

            cy.contains('label', 'I have a code inviting me to access my LPA').click();
            cy.url().should('contain', '/add-an-lpa');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/enter-access-code');
            cy.get('#f-donor-last-name').type('Smith');
            cy.get('#f-access-code').invoke('val', randomCode);
            cy.checkA11yApp();

            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/dashboard');

            cy.contains('a', 'Manage LPAs');
        })
    })
});
