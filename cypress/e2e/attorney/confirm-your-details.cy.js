import { TestMobile, TestMobile2 } from '../../support/e2e';

describe('Confirm your details', () => {
    describe('online donor', () => {
        beforeEach(() => {
            cy.visit('/fixtures/attorney?redirect=/task-list');

            cy.contains('li', 'Confirm your details').should('contain', 'Not started').click();

            cy.get('#f-phone').type(TestMobile);
            cy.contains('button', 'Save and continue').click();

            cy.get('[name="language-preference"]').check('cy', { force: true })
            cy.contains('button', 'Save and continue').click()
        });

        it('shows details', () => {
            cy.url().should('contain', '/confirm-your-details');
            cy.checkA11yApp();

            cy.contains('2 January 2000');
            cy.contains('Jessie Jones');
            cy.contains('2 RICHMOND PLACE');
            cy.contains('07700 900 000');
            cy.contains('Welsh');

            cy.contains('button', 'Continue').click();
            cy.url().should('contain', '/task-list');
        });
    });

    describe('paper donor', () => {
        beforeEach(() => {
            cy.visit('/fixtures/attorney?is-paper-donor=1&redirect=/task-list');

            cy.contains('li', 'Confirm your details').should('contain', 'Not started').click();

            cy.get('#f-phone').type(TestMobile);
            cy.contains('button', 'Save and continue').click();

            cy.get('[name="language-preference"]').check('cy', { force: true })
            cy.contains('button', 'Save and continue').click()
        });

        it('shows details', () => {
            cy.url().should('contain', '/confirm-your-details');
            cy.checkA11yApp();

            cy.contains('2 January 2000');
            cy.contains('Jessie Jones');
            cy.contains('2 RICHMOND PLACE');
            cy.contains('07700 900 000');
            cy.contains('Welsh');

            cy.contains('button', 'Continue').click();
            cy.url().should('contain', '/task-list');
        });
    });

    describe('paper donor gave phone number', () => {
        beforeEach(() => {
            cy.visit('/fixtures/attorney?is-paper-donor=1&has-phone-number=1&redirect=/task-list');

            cy.contains('li', 'Confirm your details').should('contain', 'Not started').click();

            cy.get('[name="language-preference"]').check('cy', { force: true })
            cy.contains('button', 'Save and continue').click()
        });

        it('shows details', () => {
            cy.url().should('contain', '/confirm-your-details');
            cy.checkA11yApp();

            cy.contains('h2', 'Details you have given us').next().within(() => {
                cy.contains('Welsh');
                cy.contains('Phone number').should('not.exist');
            });

            cy.contains('h2', 'Details the donor has given about you').next().within(() => {
                cy.contains('2 January 2000');
                cy.contains('Jessie Jones');
                cy.contains('2 RICHMOND PLACE');
                cy.contains('07700 900 000');
            });

            cy.contains('button', 'Continue').click();
            cy.url().should('contain', '/task-list');
        });

        it('can change the phone number', () => {
            cy.contains('.govuk-summary-list__row', 'Phone number').contains('a', 'Change').click();
            cy.get('#f-phone').clear().type(TestMobile2);
            cy.contains('button', 'Save and continue').click()

            cy.contains('h2', 'Details you have given us').next().within(() => {
                cy.contains('07700 900 111');
                cy.contains('Welsh');
            });

            cy.contains('h2', 'Details the donor has given about you').next().within(() => {
                cy.contains('2 January 2000');
                cy.contains('Jessie Jones');
                cy.contains('2 RICHMOND PLACE');
                cy.contains('Phone number').should('not.exist');
            });
        });

        it('can remove the phone number', () => {
            cy.contains('.govuk-summary-list__row', 'Phone number').contains('a', 'Change').click();
            cy.get('#f-phone').clear();
            cy.contains('button', 'Save and continue').click()

            cy.contains('h2', 'Details you have given us').next().within(() => {
                cy.contains('Enter phone number');
                cy.contains('Welsh');
            });

            cy.contains('h2', 'Details the donor has given about you').next().within(() => {
                cy.contains('2 January 2000');
                cy.contains('Jessie Jones');
                cy.contains('2 RICHMOND PLACE');
                cy.contains('Phone number').should('not.exist');
            });
        });
    });
});
