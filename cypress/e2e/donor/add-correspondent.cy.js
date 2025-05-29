import {AddressFormAssertions, DonorName} from "../../support/e2e";

describe('Add correspondent', () => {
    beforeEach(() => {
        cy.visit('/fixtures?progress=provideYourDetails&redirect=');
    });

    it('allows none', () => {
        cy.contains('M-FAKE-').click();
        cy.contains('Go to task list').click();
        cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

        cy.checkA11yApp();
        cy.contains('label', 'No').click();
        cy.contains('button', 'Save and continue').click();
        cy.contains('li', 'Add a correspondent').should('contain', 'Completed');
    });

    it('allows without address', () => {
        cy.contains('M-FAKE-').click();
        cy.contains('Go to task list').click();
        cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

        cy.checkA11yApp();
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Smith');
        cy.get('#f-email').invoke('val', 'email@example.com');
        cy.contains('label', 'No').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.contains('You’ve added a correspondent');
        cy.contains('a', 'Return to task list').click();

        cy.contains('li', 'Add a correspondent').should('contain', 'Completed');

        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=correspondent-updated&detail=${uid}`);
                    cy.contains(`{"uid":"${uid}",`);
                    cy.contains(`"firstNames":"John","lastName":"Smith","email":"email@example.com"}`);
                });
            });
    });

    it('allows with address', () => {
        cy.contains('M-FAKE-').click();
        cy.contains('Go to task list').click();
        cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

        cy.checkA11yApp();
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Smith');
        cy.get('#f-email').invoke('val', 'email@example.com');
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.contains('You’ve added a correspondent');
        cy.contains('a', 'Return to task list').click();

        cy.contains('li', 'Add a correspondent').should('contain', 'Completed');

        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=correspondent-updated&detail=${uid}`)
                    cy.contains(`{"uid":"${uid}",`);
                    cy.contains(`"firstNames":"John","lastName":"Smith","email":"email@example.com","address":{"line1":"2 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}}`);
                });
            });
    });

    it('warns when name shared with donor', () => {
        cy.contains('M-FAKE-').click();
        cy.contains('Go to task list').click();
        cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

        cy.checkA11yApp();
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.get('#f-first-names').invoke('val', DonorName.FirstNames);
        cy.get('#f-last-name').invoke('val', DonorName.LastName);
        cy.get('#f-email').invoke('val', 'email@example.com');
        cy.contains('label', 'No').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/warning');
        cy.contains('You and your correspondent have the same name. As the donor, you will automatically receive correspondence from the Office of the Public Guardian unless you nominate another person for this role.');

        cy.contains('a', 'Continue').click();
        cy.url().should('contain', '/correspondent-summary');
    });
});
