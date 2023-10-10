describe('LPA progress', () => {
    it('when nothing completed', () => {
        cy.visit('/fixtures?redirect=/progress');
        cy.checkA11yApp();

        cy.contains('li', 'You’ve signed your LPA In progress');
        cy.contains('li', 'Your certificate provider has provided their certificate Not started');
        cy.contains('li', 'Your attorneys have signed your LPA Not started');
        cy.contains('li', 'We have received your LPA Not started');
        cy.contains('li', 'Your 4-week waiting period has started Not started');
        cy.contains('li', 'Your LPA has been registered Not started');
    })

    it('when LPA submitted', () => {
        cy.visit('/fixtures?redirect=/progress&progress=confirmYourIdentityAndSignTheLpa');
        cy.checkA11yApp();

        cy.contains('li', 'You’ve signed your LPA Completed');
        cy.contains('li', 'Charlie Cooper has provided their certificate In progress');
        cy.contains('li', 'Your attorneys have signed your LPA Not started');
        cy.contains('li', 'We have received your LPA Not started');
        cy.contains('li', 'Your 4-week waiting period has started Not started');
        cy.contains('li', 'Your LPA has been registered Not started');
    })

    it('when certificate provided', () => {
        cy.visit('/fixtures?redirect=/progress&progress=signedByCertificateProvider');
        cy.checkA11yApp();

        cy.contains('li', 'You’ve signed your LPA Completed');
        cy.contains('li', 'Charlie Cooper has provided their certificate Completed');
        cy.contains('li', 'Your attorneys have signed your LPA In progress');
        cy.contains('li', 'We have received your LPA Not started');
        cy.contains('li', 'Your 4-week waiting period has started Not started');
        cy.contains('li', 'Your LPA has been registered Not started');
    })
});
