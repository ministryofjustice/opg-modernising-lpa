import { randomShareCode } from "../../support/e2e";

describe('Reuse person to notify', () => {
    beforeEach(() => {
        const sub = randomShareCode();

        cy.visit(`/fixtures?donorSub=${sub}&progress=peopleToNotifyAboutYourLpa&redirect=/task-list`);
        cy.visit(`/fixtures?donorSub=${sub}&progress=chooseYourCertificateProvider&redirect=/choose-people-to-notify`);
    });

    it('can select a previously entered person to notify', () => {
        cy.checkA11yApp();
        cy.contains('label', 'Select Jordan Jefferson').click();
        cy.contains('button', 'Continue').click();

        cy.checkA11yApp();
        cy.contains('You’ve added a person to notify');
        cy.contains('Jordan Jefferson');
    });

    it('can enter a new person to notify', () => {
        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();

        cy.url().should('include', '/enter-person-to-notify');
    });
});
