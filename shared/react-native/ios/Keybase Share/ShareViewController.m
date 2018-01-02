//
//  ShareViewController.m
//  Keybase Share
//
//  Created by Michael Maxim on 12/28/17.
//  Copyright © 2017 Keybase. All rights reserved.
//

#import "ShareViewController.h"
#import <keybase/keybase.h>

@interface ShareViewController ()

@end

@implementation ShareViewController

-(void) sendingViewController:(ItemViewController *)controller sentItem:(NSString *)retItem {
  
  // Set the configuration item's value to the returned value
  
  [item setValue:retItem];
  
  // Pop the configuration view controller to return to this one.
  
  [self popConfigurationViewController];
  
}

- (BOOL)isContentValid {
    // Do validation of contentText and/or NSExtensionContext attachments here
    return YES;
}

- (void)didSelectPost {
    // This is called after the user selects Post. Do the upload of contentText and/or NSExtensionContext attachments.
#if TESTING
    return
#endif
    
    BOOL securityAccessGroupOverride = true;
    BOOL skipLogFile = false;
    
    //NSString * home = NSHomeDirectory();
    NSFileManager* fm = [NSFileManager defaultManager];
    NSString* home = [[fm containerURLForSecurityApplicationGroupIdentifier:@"group.keybase"] path];
    
    NSString * keybasePath = [@"~/Library/Application Support/Keybase" stringByExpandingTildeInPath];
    NSString * levelDBPath = [@"~/Library/Application Support/Keybase/keybase.leveldb" stringByExpandingTildeInPath];
    NSString * chatLevelDBPath = [@"~/Library/Application Support/Keybase/keybase.chat.leveldb" stringByExpandingTildeInPath];
    NSString * logPath = [@"~/Library/Caches/Keybase" stringByExpandingTildeInPath];
    NSString * serviceLogFile = skipLogFile ? @"" : [logPath stringByAppendingString:@"/ios.log"];
  
    
    // Make keybasePath if it doesn't exist
    [fm createDirectoryAtPath:keybasePath
  withIntermediateDirectories:YES
                   attributes:nil
                        error:nil];
    NSError * err;
    KeybaseInit(home, NULL, @"prod", @(securityAccessGroupOverride), NULL, &err);
  
  KeybaseSendTextByName(@"mikem,kb_monbot", self.contentText);
    NSLog(@"CONTENT TEXT: %@", self.contentText);
  
  KeybaseShutdown();
  
    // Inform the host that we're done, so it un-blocks its UI. Note: Alternatively you could call super's -didSelectPost, which will similarly complete the extension context.
    [self.extensionContext completeRequestReturningItems:@[] completionHandler:nil];
}

SLComposeSheetConfigurationItem *item;
- (NSArray *)configurationItems {
  item = [[SLComposeSheetConfigurationItem alloc] init];
  item.title = @"Conversation";
  item.value = @"mikem";
  item.tapHandler = ^{
    ItemViewController *vie = [[ItemViewController] alloc]
    UITableViewController *viewController = [[UITableViewController alloc] initWithStyle:UITableViewStylePlain];
    [viewController tableView:<#(nonnull UITableView *)#> willBeginEditingRowAtIndexPath:<#(nonnull NSIndexPath *)#>
    [self pushConfigurationViewController:viewController];
  };
    // To add configuration options via table cells at the bottom of the sheet, return an array of SLComposeSheetConfigurationItem here.
    return @[item];
}

@end
