import { AddressFormAssertions } from "../../support/e2e";

describe('LPA type', () => {
    it('can be submitted', () => {
        cy.visit('/fixtures?redirect=/your-name');

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');

        cy.contains('button', 'Save and continue').click();

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.contains('button', 'Save and continue').click();

        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.contains('a', 'Continue').click();

        cy.get('#f-selected').check({ force: true });
        cy.contains('button', 'Save and continue').click();

        cy.get('[name="contact-language"]').check('en', { force: true })
        cy.get('[name="lpa-language"]').check('en', { force: true })

        cy.contains('button', 'Save and continue').click()

        cy.contains('a', 'Continue').click();

        cy.get('#f-lpa-type').check('property-and-affairs');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/task-list');

        cy.url().then((url) => {
            cy.visit(`http://localhost:9001/?detail-type=uid-requested&detail=${url.split('/')[4]}`);
            cy.contains(`"lpaID":"${url.split('/')[4]}"`);
        });

        cy.visit('/dashboard')

        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();
                cy.visit(`http://localhost:9001/?detail-type=application-updated&detail=${uid}`);

                cy.contains(`"uid":"${uid}"`);
                cy.contains('"type":"property-and-affairs"');
            });
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/lpa-type');

        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the type of LPA to make');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select the type of LPA to make');
    });
});
