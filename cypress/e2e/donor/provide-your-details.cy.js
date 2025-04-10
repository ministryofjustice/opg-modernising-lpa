import { AddressFormAssertions } from "../../support/e2e";

describe('Provide your details', () => {
    it('requests a UID', () => {
        cy.visit('/fixtures?redirect=');

        cy.contains('Start now').click();

        const rnd = Cypress._.random(0, 1e6);

        cy.get('#f-first-names').invoke('val', 'John' + rnd);
        cy.get('#f-last-name').invoke('val', 'Doe' + rnd);
        cy.contains('button', 'Save and continue').click();

        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1990');
        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        AddressFormAssertions.assertCanAddAddressFromSelect();

        cy.contains('a', 'Continue').click();

        cy.get('#f-selected').check({ force: true });
        cy.contains('button', 'Save and continue').click();

        cy.get('[name="contact-language"]').check('en', { force: true });
        cy.get('[name="lpa-language"]').check('en', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.contains('a', 'Continue').click();

        cy.get('#f-lpa-type').check('property-and-affairs', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url()
            .then(url => {
                const id = url.split('/')[4];

                cy.origin('http://localhost:9001', { args: { rnd, id } }, ({ rnd, id }) => {
                    cy.visit(`/?detail-type=uid-requested&detail=${id}`);
                    cy.contains('"type":"property-and-affairs"');
                    cy.contains(`"name":"John${rnd} Doe${rnd}"`);
                    cy.contains('"dob":"1990-02-01"');
                    cy.contains('"postcode":"B14 7ED"');
                });
            });
    });
});
