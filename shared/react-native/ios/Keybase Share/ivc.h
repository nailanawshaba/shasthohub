//
//  Header.h
//  Keybase
//
//  Created by Michael Maxim on 12/28/17.
//  Copyright © 2017 Keybase. All rights reserved.
//

#ifndef __ivc_h
#define __ivc_h
#import <Foundation/Foundation.h>
#import <UIKit/UIKit.h>

@class ItemViewController;

@protocol ItemViewDelegate <NSObject>

-(void)sendingViewController:(ItemViewController *) controller sentItem:(NSString *) retItem;

@end

@interface ItemViewController : UIViewController <UITextFieldDelegate>

@property (nonatomic, weak) id <ItemViewDelegate> delegate;

@end


#endif /* __ivc_h */
