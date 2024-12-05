describe('Progress', () => {
    it('when nothing completed', () => {
        cy.visit('/fixtures?redirect=/progress');
        cy.checkA11yApp();

        cy.contains('Important:').should('not.exist');

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

    it('shows a notification when going to the post office', () => {
        cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');

        cy.contains('a', 'Confirm your identity').click();
        cy.contains('button', 'Continue').click();
        cy.go(-2);
        cy.contains('a', 'Confirm your identity').click();
        cy.contains('label', 'I will confirm my identity at a Post Office').click();
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/progress');
        cy.checkA11yApp();
        cy.contains('Important:')
        cy.contains('1 notification from OPG');
        cy.contains('You have chosen to confirm your identity at a Post Office');

        cy.contains('a', 'Go to task list').click();
        cy.contains('a', 'Confirm your identity').click();
        cy.contains('label', 'to complete my Post Office identity confirmation').click();
        cy.contains('button', 'Continue').click();
        cy.contains('button', 'Continue').click();

        cy.contains('Important:').should('not.exist');
        cy.visitLpa('/progress');
    });
});
