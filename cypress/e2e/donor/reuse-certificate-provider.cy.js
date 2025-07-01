import { randomAccessCode } from "../../support/e2e";

describe('Reuse certificate provider', () => {
    before(() => {
        const sub = randomAccessCode();

        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourCertificateProvider&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=addRestrictionsToTheLpa&redirect=/task-list`);
    });

    it('selects a previously entered certificate provider', () => {
        cy.contains('li', 'Choose your certificate provider').should('contain', 'Not started').click();

        cy.contains('a', 'Continue').click();
        cy.contains('a', 'Continue').click();

        cy.contains('label', 'Select Charlie Cooper').click();
        cy.contains('button', 'Save and continue').click();

        cy.checkA11yApp();
        cy.contains('You’ve added a certificate provider');
        cy.contains('Charlie Cooper');
        cy.contains('5 RICHMOND PLACE');
    });
});
