import { AddressFormAssertions } from "../../support/e2e";

describe('Provide your details', () => {
    it('requests a UID', () => {
        cy.visit('/fixtures?redirect=');

        cy.contains('Start now').click();

        const rnd = Cypress._.random(0, 1e6);

        cy.get('#f-first-names').type('John' + rnd);
        cy.get('#f-last-name').type('Doe' + rnd);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.get('#f-can-sign').check();
        cy.contains('button', 'Continue').click();

        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.get('[name="language-preference"]').check('en')
        cy.contains('button', 'Save and continue').click()

        cy.get('#f-lpa-type').check('property-and-affairs');
        cy.contains('button', 'Continue').click();

        cy.url()
            .then(url => {
                const id = url.split('/')[4];

                cy.visit(`http://localhost:9001/?detail-type=uid-requested&detail=${id}`);
                cy.contains('"Type":"pfa"');
                cy.contains(`"name":"John${rnd} Doe${rnd}"`);
                cy.contains('"dob":"1990-02-01"');
                cy.contains('"postcode":"B14 7ED"');
            });
    });
});
