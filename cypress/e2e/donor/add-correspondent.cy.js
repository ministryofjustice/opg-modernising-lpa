import { AddressFormAssertions } from "../../support/e2e";

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
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Smith');
        cy.get('#f-email').type('email@example.com');
        cy.contains('label', 'No').click();
        cy.contains('button', 'Save and continue').click();
        cy.contains('li', 'Add a correspondent').should('contain', 'Completed');

        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.visit(`http://localhost:9001/?detail-type=correspondent-updated&detail=${uid}`);

                cy.contains(`{"uid":"${uid}",`);
                cy.contains(`"firstNames":"John","lastName":"Smith","email":"email@example.com"}`);
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
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Smith');
        cy.get('#f-email').type('email@example.com');
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.contains('li', 'Add a correspondent').should('contain', 'Completed');

        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.visit(`http://localhost:9001/?detail-type=correspondent-updated&detail=${uid}`);

                cy.contains(`{"uid":"${uid}",`);
                cy.contains(`"firstNames":"John","lastName":"Smith","email":"email@example.com","address":{"line1":"2 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}}`);
            });
    });
});
