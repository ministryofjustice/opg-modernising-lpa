import { randomAccessCode } from "../../support/e2e";

describe('Reuse person to notify', () => {
    beforeEach(() => {
        const sub = randomAccessCode();

        cy.visit(`/fixtures?donorSub=${sub}&progress=peopleToNotifyAboutYourLpa&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourCertificateProvider&redirect=/choose-people-to-notify`);
    });

    it('can select a previously entered person to notify', () => {
        cy.checkA11yApp();
        cy.contains('label', 'Select Jordan Jefferson').click();
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('People to notify about the LPA');
        cy.contains('Jordan Jefferson');
    });

    it('can enter a new person to notify', () => {
        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();

        cy.url().should('include', '/enter-person-to-notify');
    });
});
