use camino::Utf8PathBuf;
use ltk_overlay::{OverlayBuilder, EnabledMod, FantomeContent};
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 5 { return Err("usage: ame-ltk-overlay GAME OVERLAY STATE PACKAGE...".into()); }
    let mut mods = Vec::new();
    for (i, path) in args[4..].iter().enumerate() {
        let content = FantomeContent::new(std::io::BufReader::new(std::fs::File::open(path)?))?;
        mods.push(EnabledMod { id: format!("ame-{i}"), content: Box::new(content) as Box<dyn ltk_overlay::ModContentProvider>, enabled_layers: None });
    }
    let mut builder = OverlayBuilder::new(Utf8PathBuf::from(&args[1]), Utf8PathBuf::from(&args[2]), Utf8PathBuf::from(&args[3]));
    builder.set_enabled_mods(mods);
    let result = builder.build()?;
    println!("Built {} WADs", result.wads_built.len());
    Ok(())
}
