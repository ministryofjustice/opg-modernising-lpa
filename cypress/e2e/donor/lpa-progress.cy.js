describe('LPA progress', () => {
    it('when nothing completed', () => {
        cy.visit('/fixtures?redirect=/progress');
        cy.checkA11yApp();

        cy.contains('li', 'LPA paid for Not completed');
        cy.contains('li', 'Your identity confirmed Not completed');
        cy.contains('li', 'LPA signed by you Not completed');
        cy.contains('li', 'LPA certificate provided Not completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed');
        cy.contains('li', 'LPA registered by OPG Not completed');
    })

    it('when LPA submitted', () => {
        cy.visit('/fixtures?redirect=/progress&progress=signTheLpa');
        cy.checkA11yApp();

        cy.contains('li', 'LPA paid for Completed');
        cy.contains('li', 'Your identity confirmed Completed');
        cy.contains('li', 'LPA signed by you Completed');
        cy.contains('li', 'LPA certificate provided Not completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed');
        cy.contains('li', 'LPA registered by OPG Not completed');
    })

    it('when certificate provided', () => {
        cy.visit('/fixtures?redirect=/progress&progress=signedByCertificateProvider');
        cy.checkA11yApp();

        cy.contains('li', 'LPA paid for Completed');
        cy.contains('li', 'Your identity confirmed Completed');
        cy.contains('li', 'LPA signed by you Completed');
        cy.contains('li', 'LPA certificate provided Completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed');
        cy.contains('li', 'LPA registered by OPG Not completed');
    })
});
